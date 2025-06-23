//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/internal/user/adapter"
	"github.com/Wenrh2004/lark-lite-server/internal/user/application"
	"github.com/Wenrh2004/lark-lite-server/internal/user/domain"
	"github.com/Wenrh2004/lark-lite-server/internal/user/infrastructure/repository"
	adapterpkg "github.com/Wenrh2004/lark-lite-server/pkg/adapter"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/app"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/server/rpc"
	domainpkg "github.com/Wenrh2004/lark-lite-server/pkg/domain"
	repopkg "github.com/Wenrh2004/lark-lite-server/pkg/infrastruct/repository"
	"github.com/Wenrh2004/lark-lite-server/pkg/jwt"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
	"github.com/Wenrh2004/lark-lite-server/pkg/sid"
)

var infrastructureSet = wire.NewSet(
	repopkg.NewDB,
	repopkg.NewRedis,
	repository.NewTransaction,
	repository.NewRepository,
	repository.NewUserRepository,
)

var domainSet = wire.NewSet(
	domainpkg.NewService,
	domain.NewUserService,
)

var adapterSet = wire.NewSet(
	adapterpkg.NewService,
	adapter.NewUserServiceImpl,
)

var applicationSet = wire.NewSet(
	application.NewUserRPCApplication,
)

// build App
func newApp(
	// httpServer *http.Server,
	rpcServer *rpc.Server,
	conf *viper.Viper,
	// jobServer *job.Server,
	// task *server.Task,
) *app.App {
	return app.NewApp(
		// app.WithServer(httpServer),
		app.WithServer(rpcServer),
		// app.WithServer(jobServer),
		app.WithName(conf.GetString("app.name")),
	)
}

func NewWire(*viper.Viper, *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		infrastructureSet,
		domainSet,
		adapterSet,
		applicationSet,
		jwt.NewJwt,
		sid.NewSid,
		newApp,
	))
}
