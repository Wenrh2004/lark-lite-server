package nacos

import (
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/kitex-contrib/registry-nacos/v2/resolver"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
)

func NewNacosResolver(conf *viper.Viper) discovery.Resolver {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(conf.GetString("app.register.nacos.addr"), conf.GetUint64("app.register.nacos.port")),
	}

	cc := constant.ClientConfig{
		NamespaceId:         conf.GetString("app.register.nacos.namespace"),
		TimeoutMs:           conf.GetUint64("app.register.nacos.timeout"),
		NotLoadCacheAtStart: conf.GetBool("app.register.nacos.is_load_cache"),
		LogDir:              conf.GetString("app.register.nacos.log_dir"),
		CacheDir:            conf.GetString("app.register.nacos.cache_dir"),
		LogLevel:            conf.GetString("app.register.nacos.log_level"),
		Username:            conf.GetString("app.register.nacos.username"),
		Password:            conf.GetString("app.register.nacos.password"),
	}

	cli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(err)
	}

	return resolver.NewNacosResolver(cli)

}
