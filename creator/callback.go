package creator

import (
	"context"

	"github.com/gotd/td/telegram"
)

type ClientCallback = func(ctx context.Context, client *telegram.Client) error
