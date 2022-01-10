package dialog

import (
	"context"
	"strings"

	"github.com/gen2brain/dlgs"
	"github.com/go-faster/errors"
	"golang.org/x/text/message"

	tgauth "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/auth/localization"
)

var _ tgauth.UserAuthenticator = Dialog{}

// Dialog is authenticator implementation using GUI dialogs.
type Dialog struct {
	printer *message.Printer
}

// NewDialog creates new Dialog.
func NewDialog() Dialog {
	return Dialog{
		printer: localization.DefaultPrinter(),
	}
}

// WithPrinter sets localization printer.
func (d Dialog) WithPrinter(printer *message.Printer) Dialog {
	d.printer = printer
	return d
}

var errDialogClosed = errors.New("dialog closed")

// Phone implements telegram.UserAuthenticator.
func (d Dialog) Phone(ctx context.Context) (string, error) {
	r, ok, err := dlgs.Entry(
		d.printer.Sprintf(localization.PhoneDialogTitle),
		d.printer.Sprintf(localization.PhoneDialogPrompt),
		"",
	)
	if err != nil {
		return "", errors.Errorf("show dialog: %w", err)
	}
	if !ok {
		return "", errDialogClosed
	}

	return r, nil
}

// Password implements telegram.UserAuthenticator.
func (d Dialog) Password(ctx context.Context) (string, error) {
	r, ok, err := dlgs.Password(
		d.printer.Sprintf(localization.PasswordDialogTitle),
		d.printer.Sprintf(localization.PasswordDialogPrompt),
	)
	if err != nil {
		return "", errors.Errorf("show dialog: %w", err)
	}
	if !ok {
		return "", errDialogClosed
	}

	return r, nil
}

// AcceptTermsOfService implements telegram.UserAuthenticator.
func (d Dialog) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	ok, err := dlgs.Question(
		d.printer.Sprintf(localization.TOSDialogTitle),
		tos.Text,
		false,
	)
	if err != nil {
		return errors.Errorf("show dialog: %w", err)
	}
	if !ok {
		return errDialogClosed
	}

	return nil
}

// SignUp implements telegram.UserAuthenticator.
func (d Dialog) SignUp(ctx context.Context) (tgauth.UserInfo, error) {
	firstName, ok, err := dlgs.Entry(
		d.printer.Sprintf(localization.FirstNameDialogTitle),
		d.printer.Sprintf(localization.FirstNameDialogPrompt),
		"",
	)
	if err != nil {
		return tgauth.UserInfo{}, errors.Errorf("show dialog: %w", err)
	}
	if !ok {
		return tgauth.UserInfo{}, errDialogClosed
	}

	secondName, ok, err := dlgs.Entry(
		d.printer.Sprintf(localization.SecondNameDialogTitle),
		d.printer.Sprintf(localization.SecondNameDialogPrompt),
		"",
	)
	if err != nil {
		return tgauth.UserInfo{}, errors.Errorf("show dialog: %w", err)
	}
	if !ok {
		return tgauth.UserInfo{}, errDialogClosed
	}

	return tgauth.UserInfo{
		FirstName: firstName,
		LastName:  secondName,
	}, nil
}

// Code implements telegram.UserAuthenticator.
func (d Dialog) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	title := d.printer.Sprintf(localization.CodeDialogTitle)
	prompt := d.printer.Sprintf(localization.CodeDialogPrompt)
	for {
		code, ok, err := dlgs.Entry(title, prompt, "")
		if err != nil {
			return "", errors.Errorf("show dialog: %w", err)
		}
		if !ok {
			return "", errDialogClosed
		}

		code = strings.TrimSpace(code)

		type notFlashing interface {
			GetLength() int
		}

		switch v := sentCode.Type.(type) {
		case notFlashing:
			length := v.GetLength()
			if len(code) != length {
				_, err := dlgs.Error(title, d.printer.Sprintf(localization.CodeInvalidLength, length)+"\n")
				if err != nil {
					return "", errors.Errorf("write error message: %w", err)
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
