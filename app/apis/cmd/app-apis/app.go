package main

import (
	"flag"
	"fmt"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/config"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/handler"
	appmiddleware "github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/middleware"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/app-apis.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	logx.AddGlobalFields(logx.Field("service", c.Name))

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()
	server.Use(appmiddleware.SafeLog)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
