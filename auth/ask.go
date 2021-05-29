package auth

import (
	tgauth "github.com/gotd/td/telegram/auth"
)

// Ask represents parts of auth flow which
// requires user interaction.
type Ask interface {
	tgauth.CodeAuthenticator
	SignUpFlow
}

type ask struct {
	tgauth.CodeAuthenticator
	SignUpFlow
}

var _ Ask = ask{}

// BuildAsk creates new Ask.
func BuildAsk(code tgauth.CodeAuthenticator, signUp SignUpFlow) Ask {
	return ask{
		CodeAuthenticator: code,
		SignUpFlow:        signUp,
	}
}
