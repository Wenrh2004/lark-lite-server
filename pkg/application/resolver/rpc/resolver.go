package rpc

import (
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/spf13/viper"

	"github.com/Wenrh2004/lark-lite-server/pkg/application/resolver/rpc/nacos"
)

func NewResolver(conf *viper.Viper) discovery.Resolver {
	if conf.Get("app.register.nacos") != nil {
		return nacos.NewNacosResolver(conf)
	}
	return nil
}
