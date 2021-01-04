package auth

import "github.com/gotd/td/telegram"

type auth struct {
	Credentials
	Ask
}

var _ telegram.UserAuthenticator = auth{}

// Build creates new UserAuthenticator.
func Build(cred Credentials, ask Ask) telegram.UserAuthenticator {
	return auth{
		Credentials: cred,
		Ask:         ask,
	}
}
