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
// 
type OSSConfig struct {
	Region          string // OSS 地域，例如 cn-hangzhou。
	Endpoint        string // SDK 上传使用的 HTTPS Endpoint。
	Bucket          string // 保存简历的私有 Bucket 名称。
	AccessKeyID     string // 具有目标 Bucket 最小读写权限的 RAM 凭证。
	AccessKeySecret string // 与 AccessKeyID 配套的密钥。
	ObjectURLPrefix string // 稳定对象 URL 前缀，可使用 Bucket 域名或自定义域名。
}

对象 URL 前缀，可使用 Bucket 域名或自定义域名。
type ObjectStorage interface {
存储能力，便于 Logic 单元测试替换为内存 Fake。
type ObjectStorage interface {
	// Put 上传对象并返回服务端 Request ID；Request ID
	// Put 上传对象并返回服务端 Request ID；Request ID
ring, body io.Reader, size int64, contentType string) (string, error)
	// Delete 删除指定对象并返回服务端 Request
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, key string) (string, error)
S Go SDK V2 的 ObjectSto
	URL(key string) string
}

type OSSStorage struct {
	client          *oss.Client
	bucket          string
ring
}

// NewOSSStorage 
	objectURLPrefix string
}

func NewOSSStorage(config OSSConfig) (*OSSStorage, error) {
ig OSSConfig) (*OSSStorage, err
	config = trimOSSConfig(config)
	if err := validateOSSConfig(config); err != nil {
		return nil, err
	}

	sdkConfig := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.AccessKeySecret)).
		WithRegion(config.Region).
		WithEndpoint(config.Endpoint)
	return &OSSStorage{
		client:          oss.NewClient(sdkConfig),
lient:          oss.NewClient(sdkConfig),
		bucket:          config.Bucket,
		objectURLPrefix: strin
		bucket:          config.Bucket,
		objectURLPrefix: strings.TrimRight(config.ObjectURLPrefix, "/"),
	}, nil
}

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
		ForbidOverwrit
	if err != nil {
		return "", fmt.Errorf("put OSS object: %w", err)
	}
	return result.Headers.Get(oss.HeaderOssRequestID), nil
}

func (s *OSSStorage) Delete(ctx context.Context, key string) (string, error) {
	result, err := s.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(s.bucket),
		Key:    oss.Ptr(key),
	})
jectRequest{
		Bucket: oss.Ptr(s.bucket),
		Key:    oss.Ptr(key),
	})

	if err != nil {
		return "", fmt.Errorf("delete OSS object: %w", err)
	}
	return result.Headers.Get(oss.HeaderOssRequestID), nil
}


func (s *OSSStorage) URL(key string) string {
	parts := strings.Split(strings.TrimLeft(key, "/"), "/")
	for index := range parts {
		parts[index] = url.PathEscape(parts[index])
	}
	return s.objectURLPrefix + "/" + strings.Join(parts, "/")
}

func trimOSSConfig(config OSSConfig) OSSConfig {
 strings.Join(parts, "/")
}

// trimOSSConfig 去除环
	config.Region = strings.TrimSpace(config.Region)
	config.Endpoint = strings.TrimSpace(config.Endpoint)
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.AccessKeyID = strings.TrimSpace(config.AccessKeyID)
	config.AccessKeySecret = strings.TrimSpace(config.AccessKeySecret)
	config.ObjectURLPrefix = strings.TrimSpace(config.ObjectURLPrefix)
	return config
}

func validateOSSConfig(config OSSConfig) error {
eturn config
}

// validateOSSConfig 在创建 SDK 客户端前检查所有必需字段和传输协议。
func validateOSSConfig(config OSSConfig) error {
	if config.Region == "" || c
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
