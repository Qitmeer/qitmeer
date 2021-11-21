package consensus

import (
	"context"
	"github.com/Qitmeer/qitmeer/config"
)

type Context interface {
	context.Context
	GetConfig() *config.Config
}
