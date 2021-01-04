package auth

import "github.com/gotd/td/telegram"

// Ask represents parts of auth flow which
// requires user interaction.
type Ask interface {
	telegram.CodeAuthenticator
	SignUpFlow
}

type ask struct {
	telegram.CodeAuthenticator
	SignUpFlow
}

var _ Ask = ask{}

// BuildAsk creates new Ask.
func BuildAsk(code telegram.CodeAuthenticator, signUp SignUpFlow) Ask {
	return ask{
		CodeAuthenticator: code,
		SignUpFlow:        signUp,
	}
}
