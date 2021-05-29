package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	tgauth "github.com/gotd/td/telegram/auth"

	"github.com/gotd/td/tg"
)

func TestBuild(t *testing.T) {
	info := tgauth.UserInfo{
		FirstName: "FirstName",
		LastName:  "LastName",
	}
	signUp := ConstantSignUp(info)
	codeAsk := tgauth.CodeAuthenticatorFunc(func(context.Context, *tg.AuthSentCode) (string, error) {
		return "code", nil
	})

	ask := BuildAsk(
		codeAsk,
		signUp,
	)

	cred := tgauth.Constant("phone", "password", codeAsk)
	auth := Build(cred, ask)

	ctx := context.Background()
	a := require.New(t)

	code, err := auth.Code(ctx, nil)
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
