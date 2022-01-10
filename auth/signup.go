package auth

import (
	"context"

	"github.com/go-faster/errors"

	tgauth "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// SignUpFlow is abstraction for user signup setup.
type SignUpFlow interface {
	AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error
	SignUp(ctx context.Context) (tgauth.UserInfo, error)
}

// AutoAccept is noop implementation of AcceptTermsOfService call.
type AutoAccept struct{}

// AcceptTermsOfService partly implements SignUpFlow.
func (AutoAccept) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

type constantSignUp struct {
	info tgauth.UserInfo
	AutoAccept
}

func (c constantSignUp) SignUp(ctx context.Context) (tgauth.UserInfo, error) {
	return c.info, nil
}

// ConstantSignUp creates new SignUpFlow using given User info.
func ConstantSignUp(info tgauth.UserInfo) SignUpFlow {
	return constantSignUp{info: info}
}

// ErrSignUpIsNotExpected is returned, when sign up request from Telegram server
// is not expected.
var ErrSignUpIsNotExpected = errors.New("signup call is not expected")

type noSignUp struct{}

func (n noSignUp) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return &tgauth.SignUpRequired{TermsOfService: tos}
}

func (n noSignUp) SignUp(ctx context.Context) (tgauth.UserInfo, error) {
	return tgauth.UserInfo{}, ErrSignUpIsNotExpected
}

// NoSignUp creates new SignUpFlow which returns ErrSignUpIsNotExpected.
func NoSignUp() SignUpFlow {
	return noSignUp{}
}
