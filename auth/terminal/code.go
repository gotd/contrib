package terminal

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"golang.org/x/term"
	"golang.org/x/xerrors"
)

// Terminal implements UserAuthenticator.
type Terminal struct {
	*term.Terminal
}

// NewTerminal creates new Terminal.
func NewTerminal() *Terminal {
	return &Terminal{
		term.NewTerminal(os.Stdout, ""),
	}
}

func (t *Terminal) read(prompt string) (string, error) {
	t.Terminal.SetPrompt(prompt)
	defer t.Terminal.SetPrompt("")
	return t.Terminal.ReadLine()
}

func (t *Terminal) Phone(ctx context.Context) (string, error) {
	return t.read("Phone:")
}

func (t *Terminal) Password(ctx context.Context) (string, error) {
	return t.read("Password:")
}

func (t *Terminal) Code(ctx context.Context) (string, error) {
	return t.read("Code:")
}

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

func (t *Terminal) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	_, err := fmt.Fprintln(t.Terminal, tos.Text)
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
