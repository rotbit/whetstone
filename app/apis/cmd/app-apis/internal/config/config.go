package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// OSSConfig 对应 app-api
type OSSConfig struct {
	Region          string // OSS 地域。
	Endpoint        string // SDK 上传使用的 HTTPS Endpoint。
	Bucket          string // 简历文件所在的私有 Bucket。
	AccessKeyID     string // RAM 用户 AccessKey ID。
	AccessKeySecret string // RAM 用户 AccessKey Secret。
	ObjectURLPrefix string // 数据库保存的稳定对象 URL 前缀。
}

LPrefix string // 数据
type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	UserRpc      zrpc.RpcClientConf
	InterviewRpc zrpc.RpcClientConf
	QuestionRpc  zrpc.RpcClientConf
	OSS          OSSConfig // 仅 app-apis 直接访问 OSS，RPC 服务只接收稳定 URL。
	Websocket    struct {
		PublicUrl string
	}
}
