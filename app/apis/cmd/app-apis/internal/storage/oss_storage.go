package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// OSSConfig 描述 app-apis 访问阿里云 OSS 所需的最小配置。
// AccessKey 只应通过运行环境注入，不能写入仓库或输出到日志。
type OSSConfig struct {
	Region          string // OSS 地域，例如 cn-hangzhou。
	Endpoint        string // SDK 上传使用的 HTTPS Endpoint。
	Bucket          string // 保存简历的私有 Bucket 名称。
	AccessKeyID     string // 具有目标 Bucket 最小读写权限的 RAM 凭证。
	AccessKeySecret string // 与 AccessKeyID 配套的密钥。
	ObjectURLPrefix string // 稳定对象 URL 前缀，可使用 Bucket 域名或自定义域名。
}

// ObjectStorage 抽象简历上传所需的对象存储能力，便于 Logic 单元测试替换为内存 Fake。
type ObjectStorage interface {
	// Put 上传对象并返回服务端 Request ID；Request ID 仅用于日志排障。
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error)
	// Delete 删除指定对象并返回服务端 Request ID，主要用于数据库写入失败后的补偿。
	Delete(ctx context.Context, key string) (string, error)
	// URL 根据对象键构造不带临时签名参数的稳定 URL。
	URL(key string) string
}

// OSSStorage 是基于阿里云 OSS Go SDK V2 的 ObjectStorage 实现。
type OSSStorage struct {
	client          *oss.Client
	bucket          string
	objectURLPrefix string
}

// NewOSSStorage 清理并校验配置后创建 OSS 客户端。
// 配置不完整时直接返回错误，使服务启动阶段失败，而不是等到用户上传时才暴露配置问题。
func NewOSSStorage(config OSSConfig) (*OSSStorage, error) {
	config = trimOSSConfig(config)
	if err := validateOSSConfig(config); err != nil {
		return nil, err
	}

	// 使用显式静态凭证，避免 SDK 意外读取开发机上的其他默认身份。
	sdkConfig := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.AccessKeySecret)).
		WithRegion(config.Region).
		WithEndpoint(config.Endpoint)
	return &OSSStorage{
		client:          oss.NewClient(sdkConfig),
		bucket:          config.Bucket,
		objectURLPrefix: strings.TrimRight(config.ObjectURLPrefix, "/"),
	}, nil
}

// Put 以禁止覆盖模式上传对象，防止极端 UUID 冲突或错误复用对象键时覆盖历史简历。
// ContentLength 显式传给 SDK，便于 OSS 校验请求并避免流式上传大小不明确。
func (s *OSSStorage) Put(
	ctx context.Context,
	key string,
	body io.Reader,
	size int64,
	contentType string,
) (string, error) {
	result, err := s.client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket:          oss.Ptr(s.bucket),
		Key:             oss.Ptr(key),
		Body:            body,
		ContentLength:   oss.Ptr(size),
		ContentType:     oss.Ptr(contentType),
		ForbidOverwrite: oss.Ptr("true"),
	})
	if err != nil {
		return "", fmt.Errorf("put OSS object: %w", err)
	}
	return result.Headers.Get(oss.HeaderOssRequestID), nil
}

// Delete 删除补偿逻辑指定的对象；对象不存在时的具体语义由 OSS SDK 返回。
func (s *OSSStorage) Delete(ctx context.Context, key string) (string, error) {
	result, err := s.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(s.bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		return "", fmt.Errorf("delete OSS object: %w", err)
	}
	return result.Headers.Get(oss.HeaderOssRequestID), nil
}

// URL 对对象键的每个路径段分别转义，保留目录分隔符并避免空格、中文等字符破坏 URL。
// 返回值是长期稳定的对象定位地址，不包含会过期的查询签名。
func (s *OSSStorage) URL(key string) string {
	parts := strings.Split(strings.TrimLeft(key, "/"), "/")
	for index := range parts {
		parts[index] = url.PathEscape(parts[index])
	}
	return s.objectURLPrefix + "/" + strings.Join(parts, "/")
}

// trimOSSConfig 去除环境变量值两侧可能误带的空白字符。
func trimOSSConfig(config OSSConfig) OSSConfig {
	config.Region = strings.TrimSpace(config.Region)
	config.Endpoint = strings.TrimSpace(config.Endpoint)
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.AccessKeyID = strings.TrimSpace(config.AccessKeyID)
	config.AccessKeySecret = strings.TrimSpace(config.AccessKeySecret)
	config.ObjectURLPrefix = strings.TrimSpace(config.ObjectURLPrefix)
	return config
}

// validateOSSConfig 在创建 SDK 客户端前检查所有必需字段和传输协议。
func validateOSSConfig(config OSSConfig) error {
	if config.Region == "" || config.Endpoint == "" || config.Bucket == "" {
		return fmt.Errorf("OSS region, endpoint and bucket are required")
	}
	if config.AccessKeyID == "" || config.AccessKeySecret == "" {
		return fmt.Errorf("OSS access key ID and secret are required")
	}
	if err := validateHTTPSURL(config.Endpoint, "OSS endpoint"); err != nil {
		return err
	}
	return validateHTTPSURL(config.ObjectURLPrefix, "OSS object URL prefix")
}

// validateHTTPSURL 禁止明文 HTTP、URL 用户凭证、查询参数和片段。
// Endpoint 与稳定 URL 前缀都不应携带临时签名或任何敏感信息。
func validateHTTPSURL(rawURL, fieldName string) error {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" || parsedURL.User != nil {
		return fmt.Errorf("%s must be an HTTPS URL", fieldName)
	}
	if parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return fmt.Errorf("%s cannot contain query or fragment", fieldName)
	}
	return nil
}
