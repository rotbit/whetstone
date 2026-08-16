package user

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// maxResumeFileSize 与 Handler 的限制一致，Logic 层再次校验以防止被其他调用入口绕过。
	maxResumeFileSize = int64(10 << 20)
	// 上传对象统一声明为 PDF，不采信客户端提交的 Content-Type。
	pdfContentType = "application/pdf"
)

// UploadResumeLogic 负责“校验文件 -> 上传 OSS -> RPC 保存元数据”的完整业务链路。
type UploadResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUploadResumeLogic 创建带请求上下文日志的简历上传逻辑。
func NewUploadResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadResumeLogic {
	return &UploadResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UploadResume 上传当前登录用户的 PDF 简历，并返回 user-rpc 创建的简历记录。
// OSS 和 MySQL 无法组成数据库事务，因此这里通过稳定 URL 幂等和失败补偿保证最终一致性。
func (l *UploadResumeLogic) UploadResume(file io.ReadSeeker, size int64) (*types.UploadResumeResp, error) {
	// 用户 ID 只从 JWT 上下文读取，不能由 multipart 表单指定，避免替其他用户上传简历。
	userID, err := userIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := validateResumePDF(file, size); err != nil {
		return nil, err
	}

	// 服务端生成对象键，不使用用户文件名，从源头避免路径穿越和同名覆盖。
	objectKey := newResumeObjectKey(userID, time.Now().UTC(), uuid.NewString())
	requestID, err := l.svcCtx.ObjectStorage.Put(l.ctx, objectKey, file, size, pdfContentType)
	if err != nil {
		l.Errorf("upload resume to OSS failed: key=%s err=%v", objectKey, err)
		return nil, status.Error(codes.Unavailable, "上传简历失败")
	}
	l.Infof("resume uploaded to OSS: key=%s requestId=%s", objectKey, requestID)

	// 数据库只保存不带签名参数的稳定 URL，临时下载签名应在真正读取文件时按需生成。
	ossURL := l.svcCtx.ObjectStorage.URL(objectKey)
	rpcResp, err := l.saveResumeRecord(userID, ossURL)
	if err != nil {
		// InvalidArgument/AlreadyExists 能确定没有为本次对象新增记录，可以安全删除文件。
		// 其他错误可能发生在数据库提交之后、响应返回之前，此时保留对象可避免数据库悬空引用。
		if isUncertainRPCError(err) {
			l.Errorf("save resume result is uncertain: key=%s err=%v", objectKey, err)
		} else {
			l.deleteUploadedResume(objectKey)
		}
		return nil, normalizeResumeRPCError(err)
	}

	return &types.UploadResumeResp{
		ResumeId:   rpcResp.ResumeId,
		ParseState: rpcResp.ParseState,
	}, nil
}

// saveResumeRecord 调用以 OSS URL 为幂等键的 RPC；结果不确定时最多重试一次。
// 只有请求上下文仍有效才重试，避免客户端已经取消后继续占用 RPC 资源。
func (l *UploadResumeLogic) saveResumeRecord(userID int64, ossURL string) (*userclient.SaveResumeResp, error) {
	req := &userclient.SaveResumeReq{UserId: userID, OssUrl: ossURL}
	resp, err := l.svcCtx.UserRpc.SaveResume(l.ctx, req)
	if err != nil && isUncertainRPCError(err) && l.ctx.Err() == nil {
		return l.svcCtx.UserRpc.SaveResume(l.ctx, req)
	}
	return resp, err
}

// deleteUploadedResume 补偿删除尚未成功落库的 OSS 对象。
// 删除使用独立的短超时上下文，确保原请求取消后仍有机会完成清理，同时避免无限阻塞。
func (l *UploadResumeLogic) deleteUploadedResume(objectKey string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	requestID, err := l.svcCtx.ObjectStorage.Delete(ctx, objectKey)
	if err != nil {
		l.Errorf("delete uploaded resume failed: key=%s err=%v", objectKey, err)
		return
	}
	l.Infof("uploaded resume deleted: key=%s requestId=%s", objectKey, requestID)
}

// validateResumePDF 校验文件大小和 PDF 固定头，并在校验后把读取位置恢复到文件开头。
// 此处只做轻量格式识别，不承担恶意 PDF 内容扫描；更严格的安全扫描可在异步解析链路增加。
func validateResumePDF(file io.ReadSeeker, size int64) error {
	if file == nil || size <= 0 || size > maxResumeFileSize {
		return status.Error(codes.InvalidArgument, "PDF 文件不能为空且不能超过 10 MiB")
	}

	// 合法 PDF 文件以“%PDF-”开头；读取不足 5 字节同样视为格式错误。
	header := make([]byte, 5)
	if _, err := io.ReadFull(file, header); err != nil || string(header) != "%PDF-" {
		return status.Error(codes.InvalidArgument, "只支持有效的 PDF 文件")
	}
	// OSS SDK 将从当前位置读取，因此校验完成后必须回到 0 偏移。
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return status.Error(codes.InvalidArgument, "无法读取 PDF 文件")
	}
	return nil
}

// newResumeObjectKey 生成按用户和年月分层的对象键，UUID 保证同一用户可重复上传且不会覆盖历史文件。
func newResumeObjectKey(userID int64, now time.Time, objectID string) string {
	return fmt.Sprintf("resumes/%d/%04d/%02d/%s.pdf", userID, now.Year(), now.Month(), objectID)
}

// isUncertainRPCError 判断 RPC 失败后是否可能已经写入数据库。
// 当前 SaveResume 只有参数错误和 URL 归属冲突能确认未新增记录，其他错误均按不确定结果处理。
func isUncertainRPCError(err error) bool {
	switch status.Code(err) {
	case codes.InvalidArgument, codes.AlreadyExists:
		return false
	default:
		return true
	}
}

// normalizeResumeRPCError 避免把普通 Go 错误以含内部细节的 Unknown 状态直接返回给客户端。
func normalizeResumeRPCError(err error) error {
	if status.Code(err) == codes.Unknown {
		return status.Error(codes.Unavailable, "保存简历失败")
	}
	return err
}
