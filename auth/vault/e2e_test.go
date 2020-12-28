package vault_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/require"
	"github.com/tdakkota/tgcontrib/auth/vault"
	"io"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func generateToken(r io.Reader) (string, error) {
	var token [16]byte
	if _, err := io.ReadFull(r, token[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(token[:]), nil
}

type Command struct {
	Cmd   []string
	Env   []string
	Token string
}

func testEnviron() (Command, error) {
	token, err := generateToken(rand.Reader)
	if err != nil {
		return Command{}, fmt.Errorf("could not generate token: %w", err)
	}

	env := []string{
		"VAULT_DEV_LISTEN_ADDRESS=" + "0.0.0.0:8200",
		"VAULT_DEV_ROOT_TOKEN_ID=" + token,
	}

	return Command{
		Cmd:   []string{"vault", "server", "-dev"},
		Env:   env,
		Token: token,
	}, nil
}

func getDaemonAddr() (string, error) {
	var endpoint string
	if os.Getenv("DOCKER_HOST") != "" {
		endpoint = os.Getenv("DOCKER_HOST")
	} else if os.Getenv("DOCKER_URL") != "" {
		endpoint = os.Getenv("DOCKER_URL")
	} else if runtime.GOOS == "windows" {
		endpoint = "http://localhost:2375"
	} else {
		endpoint = "unix:///var/run/docker.sock"
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	var daemonAddr string
	if h := u.Hostname(); h == "localhost" || h == "0.0.0.0" {
		daemonAddr = "localhost"
	} else {
		daemonAddr = h
	}

	return daemonAddr, nil
}

func runTest(t *testing.T, addr, token string, test func(*testing.T, *api.Client)) {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 5
	bo.MaxElapsedTime = time.Minute

	var client *api.Client
	// exponential backoff-retry, because the Vault might not be ready to accept connections yet
	if err := backoff.Retry(func() (err error) {
		cfg := api.DefaultConfig()
		cfg.Address = addr
		t.Logf("Trying to connect to Vault %s", cfg.Address)

		client, err = api.NewClient(cfg)
		if err != nil {
			return
		}
		client.SetToken(token)

		_, err = client.Sys().Health()
		return
	}, bo); err != nil {
		t.Fatalf("Could not connect to Vault: %s", err)
	}

	test(t, client)
}

func testUsingDocker(test func(*testing.T, *api.Client)) func(t *testing.T) {
	return func(t *testing.T) {
		daemonAddr, err := getDaemonAddr()
		if err != nil {
			t.Fatalf("Could not parse endpoint: %s", err)
		}

		pool, err := dockertest.NewPool("")
		if err != nil {
			t.Fatalf("Could not connect to docker: %s", err)
		}

		e, err := testEnviron()
		if err != nil {
			t.Fatal(err)
		}

		resource, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "vault",
			Tag:        "latest",
			Env:        e.Env,
			Cmd:        e.Cmd,
			CapAdd: []string{
				"IPC_LOCK",
			},
		}, func(config *docker.HostConfig) {
			// set AutoRemove to true so that stopped container goes away by itself
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		})
		if err != nil {
			t.Fatalf("Could run container: %s", err)
		}

		if err := resource.Expire(60); err != nil {
			t.Fatalf("Could not set resource expiration: %s", err)
		}

		t.Cleanup(func() {
			if err := pool.Purge(resource); err != nil {
				t.Logf("Could not purge resource: %s", err)
			}
		})

		addr := fmt.Sprintf("http://%s:%s", daemonAddr, resource.GetPort("8200/tcp"))
		runTest(t, addr, e.Token, test)
	}
}

func testUsingLocalBinary(test func(*testing.T, *api.Client)) func(t *testing.T) {
	return func(t *testing.T) {
		binaryPath, ok := os.LookupEnv("TGCONTRIB_VAULT_BINARY")
		if !ok {
			path, err := exec.LookPath("vault")
			if err != nil {
				t.Fatal("Vault binary not found")
			}
			binaryPath = path
		}

		e, err := testEnviron()
		if err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(binaryPath, e.Cmd[1:]...)
		cmd.Env = append(e.Env, os.Environ()...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start %s", binaryPath)
		}
		t.Cleanup(func() {
			if err := cmd.Process.Kill(); err != nil {
				t.Logf("Could not kill process: %s", err)
			}
		})

		addr := "http://localhost:8200" // TODO:(tdakkota) set addr randomly
		runTest(t, addr, e.Token, test)
	}
}

func runUsingLocalServer(test func(*testing.T, *api.Client)) func(t *testing.T) {
	return func(t *testing.T) {
		runTest(t, os.Getenv("VAULT_ADDR"), os.Getenv("VAULT_TOKEN"), test)
	}
}

func e2etest(t *testing.T, client *api.Client) {
	a := require.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	data := []byte("mytoken")
	storage := vault.NewSessionStorage(client, "cubbyhole/test", "testtoken")
	_, err := storage.LoadSession(ctx)
	a.Error(err, "no session expected")
	a.NoError(storage.StoreSession(ctx, data))

	vaultData, err := storage.LoadSession(ctx)
	a.NoError(err)
	a.Equal(data, vaultData)

	auth := vault.NewAuth(nil, client, "cubbyhole/authtest")
	phone, password := "phone", "password"
	a.NoError(auth.SavePhone(ctx, phone))
	a.NoError(auth.SavePassword(ctx, password))

	gotPhone, err := auth.Phone(ctx)
	a.NoError(err)
	a.Equal(phone, gotPhone)

	gotPassword, err := auth.Password(ctx)
	a.NoError(err)
	a.Equal(password, gotPassword)
}

func TestE2E(t *testing.T) {
	switch os.Getenv("TGCONTRIB_VAULT_E2E_RUNNER") {
	case "local_binary":
		testUsingLocalBinary(e2etest)(t)
	case "docker":
		testUsingDocker(e2etest)(t)
	case "local_vault":
		runUsingLocalServer(e2etest)(t)
	default:
		t.Skip("Set TGCONTRIB_E2E_RUNNER to run E2E test. Possible values: docker, local_binary, local_vault")
	}
}
