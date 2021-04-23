package terminal

import (
	"context"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
	"golang.org/x/text/message"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/auth/localization"
)

var _ telegram.UserAuthenticator = (*Terminal)(nil)

// Terminal implements UserAuthenticator.
type Terminal struct {
	*term.Terminal
	printer *message.Printer
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
		printer:  localization.DefaultPrinter(),
	}
}

// OS creates new Terminal using os.Stdout and os.Stdin.
func OS() *Terminal {
	return New(os.Stdin, os.Stdout)
}

// WithPrinter sets localization printer.
func (t *Terminal) WithPrinter(printer *message.Printer) *Terminal {
	t.printer = printer
	return t
}

func (t *Terminal) read(prompt string) (string, error) {
	t.Terminal.SetPrompt(prompt)
	defer t.Terminal.SetPrompt("")
	return t.Terminal.ReadLine()
}

// Phone asks phone using terminal.
func (t *Terminal) Phone(ctx context.Context) (string, error) {
	return t.read(t.printer.Sprintf(localization.PhoneDialogPrompt) + ":")
}

// Password asks password using terminal.
func (t *Terminal) Password(ctx context.Context) (string, error) {
	return t.read(t.printer.Sprintf(localization.PasswordDialogPrompt) + ":")
}

// Code asks code using terminal.
func (t *Terminal) Code(ctx context.Context) (string, error) {
	return t.read(t.printer.Sprintf(localization.CodeDialogPrompt) + ":")
}

// SignUp asks user info for sign up.
func (t *Terminal) SignUp(ctx context.Context) (u telegram.UserInfo, err error) {
	u.FirstName, err = t.read(t.printer.Sprintf(localization.FirstNameDialogPrompt) + ":")
	if err != nil {
		return u, xerrors.Errorf("read first name: %w", err)
	}

	u.LastName, err = t.read(t.printer.Sprintf(localization.SecondNameDialogPrompt) + ":")
	if err != nil {
		return u, xerrors.Errorf("read first name: %w", err)
	}

	return u, nil
}

// AcceptTermsOfService write terms of service received from Telegram and
// asks to accept it.
func (t *Terminal) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	_, err := io.WriteString(t.Terminal, t.printer.Sprintf(localization.TOSDialogTitle)+"\n\n"+tos.Text)
	if err != nil {
		return xerrors.Errorf("write terms of service: %w", err)
	}

	t.Terminal.SetPrompt(t.printer.Sprintf(localization.TOSDialogPrompt) + "(Y/N):")
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
