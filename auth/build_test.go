package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func TestBuild(t *testing.T) {
	info := telegram.UserInfo{
		FirstName: "FirstName",
		LastName:  "LastName",
	}
	signUp := ConstantSignUp(info)
	codeAsk := telegram.CodeAuthenticatorFunc(func(ctx context.Context) (string, error) {
		return "code", nil
	})

	ask := BuildAsk(
		codeAsk,
		signUp,
	)

	cred := telegram.ConstantAuth("phone", "password", codeAsk)
	auth := Build(cred, ask)

	ctx := context.Background()
	a := require.New(t)

	code, err := auth.Code(ctx)
	a.NoError(err)
	a.Equal("code", code)

	phone, err := auth.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", phone)

	password, err := auth.Password(ctx)
	a.NoError(err)
	a.Equal("password", password)

	gotInfo, err := auth.SignUp(ctx)
	a.NoError(err)
	a.Equal(info, gotInfo)

	a.NoError(auth.AcceptTermsOfService(ctx, tg.HelpTermsOfService{}))
}
