package dialog

import (
	"context"

	"github.com/gen2brain/dlgs"
	"golang.org/x/text/message"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/auth/localization"
)

var _ telegram.UserAuthenticator = Dialog{}

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

var errDialogClosed = xerrors.New("dialog closed")

// Phone implements telegram.UserAuthenticator.
func (d Dialog) Phone(ctx context.Context) (string, error) {
	r, ok, err := dlgs.Entry(
		d.printer.Sprintf(localization.PhoneDialogTitle),
		d.printer.Sprintf(localization.PhoneDialogPrompt),
		"",
	)
	if err != nil {
		return "", xerrors.Errorf("show dialog: %w", err)
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
		return "", xerrors.Errorf("show dialog: %w", err)
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
		return xerrors.Errorf("show dialog: %w", err)
	}
	if !ok {
		return errDialogClosed
	}

	return nil
}

// SignUp implements telegram.UserAuthenticator.
func (d Dialog) SignUp(ctx context.Context) (telegram.UserInfo, error) {
	firstName, ok, err := dlgs.Entry(
		d.printer.Sprintf(localization.FirstNameDialogTitle),
		d.printer.Sprintf(localization.FirstNameDialogPrompt),
		"",
	)
	if err != nil {
		return telegram.UserInfo{}, xerrors.Errorf("show dialog: %w", err)
	}
	if !ok {
		return telegram.UserInfo{}, errDialogClosed
	}

	secondName, ok, err := dlgs.Entry(
		d.printer.Sprintf(localization.SecondNameDialogTitle),
		d.printer.Sprintf(localization.SecondNameDialogPrompt),
		"",
	)
	if err != nil {
		return telegram.UserInfo{}, xerrors.Errorf("show dialog: %w", err)
	}
	if !ok {
		return telegram.UserInfo{}, errDialogClosed
	}

	return telegram.UserInfo{
		FirstName: firstName,
		LastName:  secondName,
	}, nil
}

// Code implements telegram.UserAuthenticator.
func (d Dialog) Code(ctx context.Context, code *tg.AuthSentCode) (string, error) {
	r, ok, err := dlgs.Entry(
		d.printer.Sprintf(localization.CodeDialogTitle),
		d.printer.Sprintf(localization.CodeDialogPrompt),
		"",
	)
	if err != nil {
		return "", xerrors.Errorf("show dialog: %w", err)
	}
	if !ok {
		return "", errDialogClosed
	}

	return r, nil
}
