package job

import (
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/pkg/log"
)

type Server struct {
	rocketmq.PushConsumer
	topic string
}

func NewJob(conf *viper.Viper, logger *log.Logger) *Server {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer(conf.GetStringSlice("app.mq.path")),
		consumer.WithGroupName(conf.GetString("app.mq.consumer.group")),
	)
	if err != nil {
		panic(err)
	}

	return &Server{
		PushConsumer: c,
		topic:        conf.GetString("app.mq.topic"),
	}
}

func (j *Server) Start() {
	err := j.PushConsumer.Start()
	if err != nil {
		panic(err)
	}
}

func (j *Server) Stop() {
	err := j.PushConsumer.Shutdown()
	if err != nil {
		panic(err)
	}
}
