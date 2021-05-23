package localization

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

const (
	// PhoneDialogTitle is key for localized message.
	PhoneDialogTitle = "phone_dialog_title"
	// PhoneDialogPrompt is key for localized message.
	PhoneDialogPrompt = "phone_dialog_prompt"
	// PasswordDialogTitle is key for localized message.
	PasswordDialogTitle = "password_dialog_title"
	// PasswordDialogPrompt is key for localized message.
	PasswordDialogPrompt = "password_dialog_prompt"
	// TOSDialogTitle is key for localized message.
	TOSDialogTitle = "tos_dialog_title"
	// TOSDialogPrompt is key for localized message.
	TOSDialogPrompt = "tos_dialog_prompt"
	// FirstNameDialogTitle is key for localized message.
	FirstNameDialogTitle = "first_name_dialog_title"
	// FirstNameDialogPrompt is key for localized message.
	FirstNameDialogPrompt = "first_name_dialog_prompt"
	// SecondNameDialogTitle is key for localized message.
	SecondNameDialogTitle = "second_name_dialog_title"
	// SecondNameDialogPrompt is key for localized message.
	SecondNameDialogPrompt = "second_name_dialog_prompt"
	// CodeDialogTitle is key for localized message.
	CodeDialogTitle = "code_dialog_title"
	// CodeDialogPrompt is key for localized message.
	CodeDialogPrompt = "code_dialog_prompt"
	// CodeInvalidLength is key for localized message.
	CodeInvalidLength = "code_invalid_length"
	// CodeDoesNotMatchPattern is key for localized message.
	CodeDoesNotMatchPattern = "code_does_not_match_pattern"
)

func must(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

// Catalog returns default messages catalog.
func Catalog() *catalog.Builder {
	b := catalog.NewBuilder()
	eng := language.English

	must(
		b.SetString(eng, PhoneDialogTitle, "Your phone"),
		b.SetString(eng, PhoneDialogPrompt, "Phone"),

		b.SetString(eng, PasswordDialogTitle, "Your password"),
		b.SetString(eng, PasswordDialogPrompt, "Password"),

		b.SetString(eng, TOSDialogTitle, "Telegram requested sign up"),
		b.SetString(eng, TOSDialogPrompt, "Accept"),

		b.SetString(eng, FirstNameDialogTitle, "Your first name"),
		b.SetString(eng, FirstNameDialogPrompt, "First name"),

		b.SetString(eng, SecondNameDialogTitle, "Your last name"),
		b.SetString(eng, SecondNameDialogPrompt, "Last name"),

		b.SetString(eng, CodeDialogTitle, "Verification code"),
		b.SetString(eng, CodeDialogPrompt, "Code"),

		b.SetString(eng, CodeInvalidLength, "Code is invalid, length must be %d"),
		b.SetString(eng, CodeDoesNotMatchPattern, "Code is invalid, code must match %s"),
	)
	return b
}

// DefaultPrinter returns default localization printer.
func DefaultPrinter() *message.Printer {
	return message.NewPrinter(
		language.English,
		message.Catalog(Catalog()),
	)
}
