package userlogic

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// 标题按 Unicode 字符数限制，与 MySQL VARCHAR(128) 的字符上限保持一致。
	maxJDTitleRunes = 128
	// 正文按 UTF-8 字节数限制，为 MySQL TEXT 的 65,535 字节上限保留余量。
	maxJDContentBytes = 60 * 1024
)

// SaveJdLogic 校验并保存用户提交的原始岗位描述。
type SaveJdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewSaveJdLogic 创建带 RPC 请求上下文日志的 JD 保存逻辑。
func NewSaveJdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveJdLogic {
	return &SaveJdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SaveJd 保存一份目标岗位 JD。
// 即使 API 已校验输入，RPC 仍独立校验，防止其他内部调用方绕过边界限制。
func (l *SaveJdLogic) SaveJd(in *pb.SaveJdReq) (*pb.SaveJdResp, error) {
	if in.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 不正确")
	}

	// 去除首尾空白后再验证，空格、换行等不能构成有效标题或正文。
	title := strings.TrimSpace(in.GetTitle())
	if title == "" || utf8.RuneCountInString(title) > maxJDTitleRunes {
		return nil, status.Error(codes.InvalidArgument, "JD 标题长度必须为 1 到 128 个字符")
	}

	content := strings.TrimSpace(in.GetContent())
	if content == "" || len(content) > maxJDContentBytes {
		return nil, status.Error(codes.InvalidArgument, "JD 内容不能为空且不能超过 60 KiB")
	}

	// 当前接口只保存原始 JD，parsed_json 留给后续解析流程回填。
	result, err := l.svcCtx.JDsModel.Insert(l.ctx, &model.JD{
		UserID:  uint64(in.GetUserId()),
		Title:   title,
		Content: content,
	})
	if err != nil {
		l.Errorf("insert JD failed: %v", err)
		return nil, status.Error(codes.Internal, "保存 JD 失败")
	}

	// 将数据库自增 ID 返回给 API，供后续创建面试会话时引用。
	jdID, err := result.LastInsertId()
	if err != nil {
		l.Errorf("get inserted JD id failed: %v", err)
		return nil, status.Error(codes.Internal, "保存 JD 失败")
	}
	return &pb.SaveJdResp{JdId: jdID}, nil
}
