package auth

import (
	tgauth "github.com/gotd/td/telegram/auth"
)

type auth struct {
	Credentials
	Ask
}

var _ tgauth.UserAuthenticator = auth{}

// Build creates new UserAuthenticator.
func Build(cred Credentials, ask Ask) tgauth.UserAuthenticator {
	return auth{
		Credentials: cred,
		Ask:         ask,
	}
}
