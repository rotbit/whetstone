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

// OSSConfig 描述 app-apis 访问阿里云 OSS所需的最小配置
// AccessKey 只通过运行环境注入，不能写入仓库或输出到日志
type OSSConfig struct {
	Region          string // OSS区域，例如 cn-hangzhou
	Endpoint        string // OSS端点 SDK 上传使用的 HTTPS Endpoint。
	Bucket          string // 保存简历的私有 Bucket 名称，例如 whetstone-resumes
	AccessKeyID     string // OSS 访问密钥 ID，用于身份验证。
	AccessKeySecret string // OSS 访问密钥，用于身份验证。
	ObjectURLPrefix string // 稳定对象的URL前缀，可用Bucket域名 或自定义域名

}

// ObjectStorage 定义了对象存储的接口，用于上传、下载、删除等操作。
// 实现该接口的类型必须支持异步操作，如上传、下载、删除等。
type ObjectStorage interface {
	// Put 上传对象并返回服务端 Request ID 仅用于日志排查
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error)
	// delete 删除指定对象并返回服务端Request Id,主要用于数据库写入失败后的补偿
	Delete(ctx context.Context, key string) (string, error)
	// URL 根据对象构造不带临时签名参数的稳定URL
	URL(key string) string
}

// OSSStorage 是基于阿里云 OSS 的对象存储实现
type OSSStorage struct {
	client          *oss.Client
	bucket          string
	objectURLPrefix string
}

// NewOSSStorage 创建OSS客户端
func NewOSSStorage(config OSSConfig) (*OSSStorage, error) {
	//配置不完整时直接返回错误，使服务启动阶段失败，而不是等到用户上传时才暴露配置问题。
	config = TrimOSSConfig(config) // 移除配置中的空格
	//校验配置
	if err := validateOSSConfig(config); err != nil {
		return nil, err
	}

	// 使用显示静态凭证，避免SDK 意外读取开发机上的其他默认身份
	sdkConfig := oss.LoadDefaultConfig().WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.AccessKeySecret)).
		WithRegion(config.Region).
		WithEndpoint(config.Endpoint)
	return &OSSStorage{
		client:          oss.NewClient(sdkConfig),
		bucket:          config.Bucket,
		objectURLPrefix: strings.TrimRight(config.ObjectURLPrefix, "/"),
	}, nil
}

// Put 以禁止覆盖模式上传对象，防止极端uuid 冲突或 错误复用对象键时导致覆盖历史简历。
// ContentLength 显示传给SDK, 便于OSS 校验请求并避免流式上传大小不明确
func (s *OSSStorage) Put(ctx context.Context, key string, body io.Reader, size int64,
	contentType string) (string, error) {
	result, err := s.client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket:          oss.Ptr(s.bucket),
		Key:             oss.Ptr(key),
		Body:            body,
		ContentLength:   oss.Ptr(size),
		ContentType:     oss.Ptr(contentType),
		ForbidOverwrite: oss.Ptr("true"),
	})
	if err != nil {
		return "", fmt.Errorf("put oss object failed: %w", err)
	}
	return result.Headers.Get(oss.HeaderOssRequestID), nil
}

// Delete 删除补偿逻辑指定的对象：对象不存在时的具体语义由OSS SDK 返回
func (s *OSSStorage) Delete(ctx context.Context, key string) (string, error) {
	result, err := s.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(s.bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		return "", fmt.Errorf("delete oss object failed: %w", err)
	}
	return result.Headers.Get(oss.HeaderOssRequestID), nil
}

// URL 对对象键的每个路径段分别转义， 保留目录分隔符并避免空格、中文等字符破坏URL。
// URL 将 OSS 对象的 key（如 "resumes/2024/08/我的简历.pdf"）转换为
// 一个完整的、不带签名参数的稳定访问链接。
// 生成的链接格式：s.objectURLPrefix + "/" + 转义后的路径
// 主要用于存入数据库，供前端后续访问或下载。
func (s *OSSStorage) URL(key string) string {
	// 1. 去除路径开头的斜杠，并按 "/" 分割成切片
	//    例如 "resumes/2024/简历.pdf" -> ["resumes", "2024", "简历.pdf"]
	parts := strings.Split(strings.TrimLeft(key, "/"), "/")

	// 2. 对每一级路径进行 URL 转义，确保特殊字符（如中文、空格、()、# 等）不会破坏 URL
	//    例如 "简历.pdf" -> "%E7%AE%80%E5%8E%86.pdf"
	for index := range parts {
		parts[index] = url.PathEscape(parts[index])
	}
	// 3. 将转义后的路径用 "/" 重新拼接，并与前缀组合成完整 URL
	//    最终类似 "https://cdn.company.com/resumes/2024/%E7%AE%80%E5%8E%86.pdf"
	return s.objectURLPrefix + "/" + strings.Join(parts, "/")
}

// TrimOSSConfig 移除环境变量两侧的空格，确保配置项值没有空格
func TrimOSSConfig(config OSSConfig) OSSConfig {
	config.Region = strings.TrimSpace(config.Region)
	config.Endpoint = strings.TrimSpace(config.Endpoint)
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.AccessKeyID = strings.TrimSpace(config.AccessKeyID)
	config.AccessKeySecret = strings.TrimSpace(config.AccessKeySecret)
	config.ObjectURLPrefix = strings.TrimSpace(config.ObjectURLPrefix)
	return config
}

// validateOSSSConfig 在创建 SDK 客户端前检查所有必须字段 和 传输协议
func validateOSSConfig(config OSSConfig) error {
	if config.Region == "" || config.Endpoint == "" || config.Bucket == "" {
		return fmt.Errorf("OSS region, endpoint and bucket are required")
	}
	if config.AccessKeyID == "" || config.AccessKeySecret == "" {
		return fmt.Errorf("OSS accessKeyID and accessKeySecret are required")
	}
	if err := validateHTTPSURL(config.Endpoint, "OSS endpoint"); err != nil {
		return err
	}
	return validateHTTPSURL(config.ObjectURLPrefix, "OSS objectURLPrefix")
}

// validateHTTPSURL禁止明文 HTTP、URL 用户凭证 、查询参数 和片段。
// Endpoint 与稳定 URl 前缀都不应该携带临时签名 或 任何敏感信息
func validateHTTPSURL(rawURL, fieldName string) error {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" || parsedURL.User != nil {
		return fmt.Errorf("invalid %s: %w", fieldName, err)
	}
	if parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return fmt.Errorf("invalid %s: %w", fieldName, err)
	}
	return nil
}

/*
// validateHTTPSURL 校验一个字符串是否为合法的 HTTPS URL，
// 且禁止包含用户凭证、查询参数和片段标识符。
// 常用于 OSS Endpoint、自定义域名前缀等必须安全、干净的配置项。
// 参数：
//   rawURL   - 待校验的原始 URL 字符串（可能包含前后空格，需事先 Trim）
//   fieldName - 字段名称，用于错误信息提示（如 "Endpoint"）
// 返回：
//   error - 校验失败时返回描述性错误；成功返回 nil
func validateHTTPSURL(rawURL, fieldName string) error {
    // 1. 解析 URL（必须为绝对 URI，即包含协议头）
    parsedURL, err := url.ParseRequestURI(rawURL)

    // 2. 检查：解析失败 或 非 HTTPS 或 主机为空 或 包含用户信息
    //    注意：err != nil 时，parsedURL 可能为 nil，所以放在 || 的第一位（短路求值）
    if err != nil ||
       parsedURL.Scheme != "https" ||
       parsedURL.Host == "" ||
       parsedURL.User != nil {
        // 这里使用 %w 包裹原始错误，保留错误链，便于上层用 errors.Is 判断
        return fmt.Errorf("invalid %s: %w", fieldName, err)
    }

    // 3. 检查是否携带了查询参数（?xxx）或片段（#xxx）
    //    这些内容通常用于临时签名或前端锚点，不应作为永久配置存储
    if parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
        return fmt.Errorf("invalid %s: %w", fieldName, fmt.Errorf("query or fragment not allowed"))
    }

    // 4. 所有检查通过
    return nil
}
*/
