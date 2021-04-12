package terminal

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTerminal(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)

	var in, out bytes.Buffer
	term := New(&in, &out)
	test := func(output, input string, call func(t *Terminal) (string, error)) {
		in.WriteString(input + "\r")
		phone, err := call(term)
		a.NoError(err)
		a.Equal(input, phone)
		a.Equal(output+":"+input, strings.TrimSpace(out.String()))
		out.Reset()
	}

	test("Phone", "abc", func(t *Terminal) (string, error) {
		return t.Phone(ctx)
	})
	test("Password", "abc", func(t *Terminal) (string, error) {
		return t.Password(ctx)
	})
	test("Code", "abc", func(t *Terminal) (string, error) {
		return t.Code(ctx)
	})
}
