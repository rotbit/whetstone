package main

import (
	"flag"
	"fmt"

	"github.com/yourname/whetstone/app/question/rpc/internal/config"
	questionServer "github.com/yourname/whetstone/app/question/rpc/internal/server/question"
	"github.com/yourname/whetstone/app/question/rpc/internal/svc"
	"github.com/yourname/whetstone/app/question/rpc/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/question.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	logx.AddGlobalFields(logx.Field("service", c.Name))
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterQuestionServer(grpcServer, questionServer.NewQuestionServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
