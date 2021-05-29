package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	tgauth "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

func TestConstant(t *testing.T) {
	firstName := "first"
	lastName := "last"
	info := tgauth.UserInfo{
		FirstName: firstName,
		LastName:  lastName,
	}

	signUp := ConstantSignUp(info)
	ctx := context.Background()

	a := require.New(t)
	a.NoError(signUp.AcceptTermsOfService(ctx, tg.HelpTermsOfService{}))
	gotInfo, err := signUp.SignUp(ctx)
	a.NoError(err)
	a.Equal(info, gotInfo)
}

func TestNoSignUp(t *testing.T) {
	signUp := NoSignUp()
	ctx := context.Background()

	a := require.New(t)
	a.Error(signUp.AcceptTermsOfService(ctx, tg.HelpTermsOfService{}))
	_, err := signUp.SignUp(ctx)
	a.Error(err)
}
