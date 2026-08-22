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

	"github.com/zeromicro/go-zero/core/logx"
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
}

// validateResumePDF 校验简历文件大小 和PDF 固定头，并在校验后把读取位置恢复到文件开头
// 此处只做轻量格式识别， 不承担恶意PDF 内容扫描：更严格的安全扫描可在异步解析链路增加
func validateResumePDF(file io.ReadeSeeker, size int64) error {
	if file == nil || size <= 0 || size > maxResumeFileSize {
		return status.Error(codes.InvalidArgument, "PDF 文件不能为空且不能超过10 MiB")
	}

	// 合法 PDF 文件以“PDF-”开头: 读取不足 5 字节同样视为格式错误。
	header := make([]byte, 5)
	if _, err := file.ReadFull(file, header); err != nil || string(header) != "PDF-" {
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
