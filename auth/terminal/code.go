package terminal

import (
	"context"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// Terminal implements UserAuthenticator.
type Terminal struct {
	*term.Terminal
}

// New creates new Terminal.
func New(in io.Reader, out io.Writer) *Terminal {
	rw := struct {
		io.Reader
		io.Writer
	}{
		Reader: in,
		Writer: out,
	}
	return &Terminal{
		Terminal: term.NewTerminal(rw, ""),
	}
}

// OS creates new Terminal using os.Stdout and os.Stdin.
func OS() *Terminal {
	return New(os.Stdin, os.Stdout)
}

func (t *Terminal) read(prompt string) (string, error) {
	t.Terminal.SetPrompt(prompt)
	defer t.Terminal.SetPrompt("")
	return t.Terminal.ReadLine()
}

// Phone asks phone using terminal.
func (t *Terminal) Phone(ctx context.Context) (string, error) {
	return t.read("Phone:")
}

// Password asks password using terminal.
func (t *Terminal) Password(ctx context.Context) (string, error) {
	return t.read("Password:")
}

// Code asks code using terminal.
func (t *Terminal) Code(ctx context.Context) (string, error) {
	return t.read("Code:")
}

// SignUp asks user info for sign up.
func (t *Terminal) SignUp(ctx context.Context) (u telegram.UserInfo, err error) {
	u.FirstName, err = t.read("First name:")
	if err != nil {
		return u, xerrors.Errorf("read first name: %w", err)
	}

	u.LastName, err = t.read("Last name:")
	if err != nil {
		return u, xerrors.Errorf("read first name: %w", err)
	}

	return u, nil
}

// AcceptTermsOfService write terms of service received from Telegram and
// asks to accept it.
func (t *Terminal) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	_, err := io.WriteString(t.Terminal, "Telegram requested sign up, user not found.\n\n"+tos.Text)
	if err != nil {
		return xerrors.Errorf("write terms of service: %w", err)
	}

	t.Terminal.SetPrompt("Accept(Y/N):")
	defer t.Terminal.SetPrompt("")

loop:
	y, err := t.Terminal.ReadLine()
	if err != nil {
		return xerrors.Errorf("read answer: %w", err)
	}
	switch strings.ToLower(y) {
	case "y", "yes":
		return nil
	case "n", "no":
		return xerrors.New("user answer is no")
	default:
		goto loop
	}
}
