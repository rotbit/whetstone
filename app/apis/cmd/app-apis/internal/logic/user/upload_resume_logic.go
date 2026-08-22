package user

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// maxResumeFileSize 最大简历文件大小 10MiB  logic层再次校验， 以防止被其他调用入口绕过。
	maxResumeFileSize = 10 * 1024 * 1024
	// 上传对象统一声明为PDF 文件， 不采信客户端提交的Content-Type。
	pdfContentType = "application/pdf"
)

type UploadResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadResumeLogic {
	return &UploadResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UploadResume 上传简历
// OSS 和 MySQL 无法组成数据库事务，因此这里通过稳定的URL幂等 和失败补偿保证最终的一致性
func (l *UploadResumeLogic) UploadResume(file io.ReadSeeker, size int64) (resp *types.UploadResumeResp, err error) {
	// 用户ID 只从JWT 上下文读取， 不能由 multipart 表单指定， 避免替其他用户上传简历。
	userID, err := userIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := validateResumePDF(file, size); err != nil {
		return nil, err
	}

	// 生成对象键，不使用用户文件名， 从源头避免路径穿越 和 同名覆盖。 uuid.NewString() 生成一个随机的UUID 字符串
	objectKey := newResumeObjectKey(userID, time.Now().UTC(), uuid.NewString())
	requestID, err := l.svcCtx.ObjectStorage.Put(l.ctx, objectKey,file, size, pdfContentType)
	if err != nil {
		l.Errorf("upload resume to OSS failed, key:%s err=%v", objectKey, err)
		return  nil, status.Error(codes.Unavailable, "上传简历失败")
	}
	l.Infof("resume uploaded to OSS: key=%s, requestID=%s", objectKey, requestID)

	// 数据库只保存不带签名参数的稳定 URL，临时下载签名应在真正读取文件时按需生成
	ossURL := l.svcCtx.ObjectStorage.URL(objectKey)
	rpcResp, err := l.saveResumeRecord(userID, ossURL) 
	if err != nil {
		// InvalidArgument/AlreadyExists 能确定没有为本次对象新增记录，可以安全删除文件
		// 其他错误可能发生在数据库提交之后、响应返回之前，此时保留对象可避免数据库悬空引用
		if isRetryableError(err) { // oss 未上传成功
			l.Errorf("save resume result is uncertain: key=%s, err=%v",objectKey, err)
		} else {
			l.deleteUploadedResume(objectKey)
		}
		return nil, normalizeResumeRPCError(err)
	}
	return &types.UploadResumeResp {
		ResumeId: rpcResp.ResumeId,
		ParseState: rpcResp.ParseState,
	},nil
}

// validateResumePDF 校验简历文件大小 和PDF 固定头，并在校验后把读取位置恢复到文件开头
// 此处只做轻量格式识别， 不承担恶意PDF 内容扫描：更严格的安全扫描可在异步解析链路增加
func validateResumePDF(file io.ReadSeeker, size int64) error {
	if file == nil || size <= 0 || size > maxResumeFileSize {
		return status.Error(codes.InvalidArgument, "PDF 文件不能为空且不能超过10 MiB")
	}

	// 合法 PDF 文件以“PDF-”开头: 读取不足 5 字节同样视为格式错误。
	header := make([]byte, 5)
	if _, err := io.ReadFull(file, header); err != nil || string(header) != "PDF-" {
		return status.Error(codes.InvalidArgument, "只支持有效的 PDF 文件")
	}
	// OSS SDK 将从当前位置读取，因此校验完后必须回到 0 偏移
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return status.Error(codes.InvalidArgument, "无法读取PDF文件")
	}
	return nil
}

// newResumeObjectKey 生成按用户和年月分层的对象键, UUID 保证统一用户可重复上传且不会覆盖历史文件。
func newResumeObjectKey(userID int64, now time.Time, objectID string) string {
	return fmt.Sprintf("resumes/%d/%04d/%02d/%s.pdf", userID, now.Year(), now.Month(), objectID)
}

// saveResumeRecord 把已上传的简历 URL 落库到 user-rpc。
// 调用 user-rpc.SaveResume 可能在网络层失败而服务端实际已成功写入，
// 此时按 err 的 gRPC code 分类处理：
//   - 参数错误(InvalidArgument) 或 简历归属冲突(AlreadyExists)：重试无意义，直接返回
//   - 其他错误(Unavailable/Deadline/Internal...)：结果不确定，但 user-rpc 用 oss_url
//     作幂等键(先查后插 + 唯一索引兜底)，所以可以安全地重试一次拿稳态结果
// 重试前还检查 ctx.Err()，避免客户端已断开后还白白占用 RPC 资源。
func (l *UploadResumeLogic) saveResumeRecord(userID int64, ossURL string) (*userclient.SaveResumeResp, error) {
	req := &userclient.SaveResumeReq{UserId: userID, OssUrl: ossURL}
	resp, err := l.svcCtx.UserRpc.SaveResume(l.ctx, req)
	if err != nil && isRetryableError(err) && l.ctx.Err() == nil {
		// user-rpc.SaveResume 幂等：再调一次拿到的是稳定结果(已存在的 resume_id 或新插入)
		return l.svcCtx.UserRpc.SaveResume(l.ctx, req)
	}
	return resp, err
}

// isRetryableError 判断一次 RPC 失败后是否值得重试。
// 不可重试的语义错误(参数不对/资源归属冲突)返回 false；网络/服务端不确定状态
// 的错误返回 true，依赖被调服务的幂等设计保证重试不会产生副作用。
func isRetryableError(err error) bool {
	switch status.Code(err) {
	case codes.InvalidArgument, codes.AlreadyExists:
		return false
	default:
		return true
	}
}

// deleteUploadedResume 补偿删除尚未成功落库的 OSS 对象。
//删除功能 这使用独立的短超时上下文，确保原来请求取消后仍有机会完成清理，同时避免无限阻塞
func (l *UploadResumeLogic) deleteUploadedResume(objectKey string) {
	ctx, cancel := context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()
	requestID, err := l.svcCtx.ObjectStorage.Delete(ctx,objectKey)
	if err != nil {
		l.Errorf("delete uploaded resume failed: key=%s, err=%v",objectKey, err)
		return
	}
	l.Infof("uploaded resume delete success: key=%s, requestID=%s", objectKey, requestID)
}

// normalizeResumeRPCError 避免把普通 go 错误以含内部细节的 Unknown 状态直接返回给客户端。
func normalizeResumeRPCError(err error) error {
	if status.Code(err) == codes.Unknown {
		return status.Error(codes.Internal,"保存简历失败")
	}
	return err
}