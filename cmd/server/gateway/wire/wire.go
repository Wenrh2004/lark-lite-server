//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/internal/user/adapter"
	"github.com/Wenrh2004/lark-lite-server/internal/user/application"
	"github.com/Wenrh2004/lark-lite-server/internal/user/common"
	adapterpkg "github.com/Wenrh2004/lark-lite-server/pkg/adapter"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/app"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/resolver/rpc"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/server/http"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

var infrastructureSet = wire.NewSet(
	adapterpkg.NewService,
	rpc.NewResolver,
	common.NewUserClient,
)

var adapterSet = wire.NewSet(
	adapter.NewUserHandler,
)

var applicationSet = wire.NewSet(
	application.NewUserHTTPApplication,
)

// build App
func newApp(
	httpServer *http.Server,
	// 	rpcServer *rpc.Server,
	conf *viper.Viper,
	// jobServer *job.Server,
	// task *server.Task,
) *app.App {
	return app.NewApp(
		app.WithServer(httpServer),
		// app.WithServer(rpcServer),
		// app.WithServer(jobServer),
		app.WithName(conf.GetString("app.name")),
	)
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		infrastructureSet,
		adapterSet,
		applicationSet,
		newApp,
	))
}
