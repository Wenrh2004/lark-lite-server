package application

import (
	"github.com/apache/rocketmq-client-go/v2/consumer"
	kitexregistry "github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/server"
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/common/kitex_gen/file/fileservice"
	"github.com/Wenrh2004/lark-lite-server/internal/file/adapter"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/server/job"
	"github.com/Wenrh2004/lark-lite-server/pkg/application/server/rpc"
	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

func NewRPCApplication(logger *log.Logger, r kitexregistry.Registry, srv *adapter.FileService) *rpc.Server {
	s := fileservice.NewServer(srv, server.WithRegistry(r))
	return rpc.NewServer(s, logger)
}

func NewJobApplication(conf *viper.Viper, logger *log.Logger, fs *adapter.FileJob) *job.Server {
	j := job.NewJob(conf, logger)
	if err := j.Subscribe(conf.GetString("app.mq.topic"), consumer.MessageSelector{}, fs.UploadFailed); err != nil {
		panic(err)
	}
	return j
}
