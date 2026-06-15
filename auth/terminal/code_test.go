package terminal

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/auth/localization"
)

func TestTerminal(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)

	var in, out bytes.Buffer
	// A bytes.Buffer is not a tty, so New falls back to bufio.Reader: the
	// prompt is printed but input is not echoed back.
	term := New(&in, &out).WithPrinter(message.NewPrinter(language.English))
	test := func(output, input string, call func(t *Terminal) (string, error)) {
		in.WriteString(input)
		in.WriteString("\n")
		got, err := call(term)
		a.NoError(err)
		a.Equal(input, got)
		a.Equal(output+":", out.String())
		out.Reset()
	}

	input := "abc"
	test(localization.PhoneDialogPrompt, input, func(t *Terminal) (string, error) {
		return t.Phone(ctx)
	})
	test(localization.PasswordDialogPrompt, input, func(t *Terminal) (string, error) {
		return t.Password(ctx)
	})
	test(localization.CodeDialogPrompt, input, func(t *Terminal) (string, error) {
		return t.Code(ctx, &tg.AuthSentCode{
			Type: &tg.AuthSentCodeTypeApp{
				Length: len(input),
			},
		})
	})
}

// TestTerminalNoNewline ensures the last line is accepted even without a
// trailing newline (e.g. EOF on a pipe).
func TestTerminalNoNewline(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)

	var in, out bytes.Buffer
	in.WriteString("12345")
	term := New(&in, &out).WithPrinter(message.NewPrinter(language.English))

	got, err := term.Phone(ctx)
	a.NoError(err)
	a.Equal("12345", got)
}
