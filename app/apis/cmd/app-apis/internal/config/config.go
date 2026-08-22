package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 汇总 app-apis 的HTTP配置、鉴权、RPC、websocket配置段
// OSS 配置通过环境变量直接读取，不经过 YAML 解析
type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	UserRpc      zrpc.RpcClientConf
	InterviewRpc zrpc.RpcClientConf
	QuestionRpc  zrpc.RpcClientConf
	Websocket    struct {
		PublicUrl string
	}
}
