package terminal

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/gotd/contrib/auth/localization"
)

func TestTerminal(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)

	var in, out bytes.Buffer
	term := New(&in, &out).WithPrinter(message.NewPrinter(language.English))
	test := func(output, input string, call func(t *Terminal) (string, error)) {
		in.WriteString(input + "\r")
		phone, err := call(term)
		a.NoError(err)
		a.Equal(input, phone)
		a.Equal(output+":"+input, strings.TrimSpace(out.String()))
		out.Reset()
	}

	test(localization.PhoneDialogPrompt, "abc", func(t *Terminal) (string, error) {
		return t.Phone(ctx)
	})
	test(localization.PasswordDialogPrompt, "abc", func(t *Terminal) (string, error) {
		return t.Password(ctx)
	})
	test(localization.CodeDialogPrompt, "abc", func(t *Terminal) (string, error) {
		return t.Code(ctx)
	})
}
