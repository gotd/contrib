package auth

// CredentialNotFoundError should be returned, when
// credential not found.
type CredentialNotFoundError struct {
	Which CredentialType
}

func (c *CredentialNotFoundError) Error() string {
	return "credential " + c.Which.String() + " not found"
}

func (CredentialNotFoundError) Is(err error) bool {
	_, ok := err.(*CredentialNotFoundError)
	return ok
}

// CredentialType represents user credential type.
type CredentialType string

func (c CredentialType) String() string {
	return string(c)
}

const (
	Phone    CredentialType = "phone"
	Password CredentialType = "password"
	Code     CredentialType = "code"
	Info     CredentialType = "userinfo"
)
