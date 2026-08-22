package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 汇总 app-apis 的HTTP配置、鉴权、RPC、OSS、websocket配置段
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
	OSS OSSConfig // OSS 配置 仅app-apis直接访问OSS，RPC  服务只接收稳定的OSS URL
}

// OSSConfig 对应app-apis的oss配置段，由go-zero在启动环境时从环境变量展开
type OSSConfig struct {
	Region          string //OSS区域
	Endpoint        string // SDK 上传使用的HTTPS Endpoint
	Bucket          string // OSS 桶名 用于存储简历文件
	AccessKeyID     string // OSS 访问密钥ID
	AccessKeySecret string // OSS 密钥
	ObjectURLPrefix string // OSS 对象URL前缀 用于存储简历文件
}
