//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/internal/file/adapter"
	"github.com/Wenrh2004/lark-lite-server/internal/file/application"
	"github.com/Wenrh2004/lark-lite-server/internal/file/domain"
	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/producer"
	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/repository"
	"github.com/Wenrh2004/lark-lite-server/internal/file/infrastructure/third/oss"
	adapterpkg "github.com/Wenrh2004/lark-lite-server/pkg/adapter"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/app"
	rpcpkg "github.com/Wenrh2004/lark-lite-server/pkg/application/register/rpc"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/server/job"
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
	repository.NewFileRepository,
	producer.NewProducer,
	oss.NewService,
)

var domainSet = wire.NewSet(
	domainpkg.NewService,
	domain.NewFileService,
)

var adapterSet = wire.NewSet(
	adapterpkg.NewService,
	adapter.NewFileService,
	adapter.NewFileJob,
)

var applicationSet = wire.NewSet(
	rpcpkg.NewRegister,
	application.NewRPCApplication,
	application.NewJobApplication,
)

// build App
func newApp(
	// httpServer *http.Server,
	rpcServer *rpc.Server,
	conf *viper.Viper,
	jobServer *job.Server,
	// task *server.Task,
) *app.App {
	return app.NewApp(
		// app.WithServer(httpServer),
		app.WithServer(rpcServer),
		app.WithServer(jobServer),
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
