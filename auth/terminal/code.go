package terminal

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/term"
	"golang.org/x/text/message"

	tgauth "github.com/gotd/td/telegram/auth"

	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/auth/localization"
)

var _ tgauth.UserAuthenticator = (*Terminal)(nil)

// Terminal implements UserAuthenticator.
type Terminal struct {
	// terminal is used when the input is an interactive terminal (tty).
	terminal *term.Terminal
	// reader is the fallback used when the input is not a tty (e.g. a pipe,
	// a regular file or a buffer).
	reader  *bufio.Reader
	out     io.Writer
	printer *message.Printer
}

// New creates new Terminal.
//
// If in is an interactive terminal an interactive *term.Terminal is used,
// otherwise it falls back to a bufio.Reader, so non-tty inputs such as pipes,
// files or buffers are handled gracefully.
func New(in io.Reader, out io.Writer) *Terminal {
	t := &Terminal{
		out:     out,
		printer: localization.DefaultPrinter(),
	}
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		rw := struct {
			io.Reader
			io.Writer
		}{
			Reader: in,
			Writer: out,
		}
		t.terminal = term.NewTerminal(rw, "")
	} else {
		t.reader = bufio.NewReader(in)
	}
	return t
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

// write writes s to the interactive terminal or, in the fallback mode, to the
// output writer.
func (t *Terminal) write(s string) error {
	w := t.out
	if t.terminal != nil {
		w = t.terminal
	}
	_, err := io.WriteString(w, s)
	return err
}

// read prints the prompt and reads a single line of input.
func (t *Terminal) read(prompt string) (string, error) {
	if t.terminal != nil {
		t.terminal.SetPrompt(prompt)
		defer t.terminal.SetPrompt("")
		return t.terminal.ReadLine()
	}

	// Fallback for non-terminal input: print the prompt ourselves, since there
	// is no terminal to render it, and read a line with bufio.
	if err := t.write(prompt); err != nil {
		return "", err
	}
	line, err := t.reader.ReadString('\n')
	if err != nil {
		// Accept the last line even if it is not terminated by a newline.
		if !errors.Is(err, io.EOF) || line == "" {
			return "", err
		}
	}
	return strings.TrimRight(line, "\r\n"), nil
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
func (t *Terminal) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	prompt := t.printer.Sprintf(localization.CodeDialogPrompt)
	for {
		code, err := t.read(prompt + ":")
		if err != nil {
			return "", err
		}
		code = strings.TrimSpace(code)

		type notFlashing interface {
			GetLength() int
		}

		switch v := sentCode.Type.(type) {
		case notFlashing:
			length := v.GetLength()
			if len(code) != length {
				if err := t.write(t.printer.Sprintf(localization.CodeInvalidLength, length) + "\n"); err != nil {
					return "", errors.Wrap(err, "write error message")
				}
				continue
			}

			return code, nil
		// TODO: add tg.AuthSentCodeTypeFlashCall support
		default:
			return code, nil
		}
	}
}

// SignUp asks user info for sign up.
func (t *Terminal) SignUp(ctx context.Context) (u tgauth.UserInfo, err error) {
	u.FirstName, err = t.read(t.printer.Sprintf(localization.FirstNameDialogPrompt) + ":")
	if err != nil {
		return u, errors.Wrap(err, "read first name")
	}

	u.LastName, err = t.read(t.printer.Sprintf(localization.SecondNameDialogPrompt) + ":")
	if err != nil {
		return u, errors.Wrap(err, "read first name")
	}

	return u, nil
}

// AcceptTermsOfService write terms of service received from Telegram and
// asks to accept it.
func (t *Terminal) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	if err := t.write(t.printer.Sprintf(localization.TOSDialogTitle) + "\n\n" + tos.Text); err != nil {
		return errors.Wrap(err, "write terms of service")
	}

	prompt := t.printer.Sprintf(localization.TOSDialogPrompt) + "(Y/N):"
	for {
		y, err := t.read(prompt)
		if err != nil {
			return errors.Wrap(err, "read answer")
		}
		switch strings.ToLower(strings.TrimSpace(y)) {
		case "y", "yes":
			return nil
		case "n", "no":
			return errors.New("user answer is no")
		}
	}
}
