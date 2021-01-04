package auth

import "context"

// Credentials represents Telegram user credentials.
type Credentials interface {
	Phone(ctx context.Context) (string, error)
	Password(ctx context.Context) (string, error)
}
