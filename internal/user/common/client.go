package common

import (
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/common/kitex_gen/user/userservice"
)

func NewUserClient(conf *viper.Viper, r discovery.Resolver) userservice.Client {
	cli, err := userservice.NewClient(
		conf.GetString("user.service_name"),
		client.WithResolver(r),
		client.WithRPCTimeout(time.Second*3),
	)
	if err != nil {
		panic(err)
	}
	return cli
}
