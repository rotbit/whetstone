package userlogic

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 新记录先进入 parsing，后续解析任务再更新为 done 或 failed。
	defaultResumeParseState = "parsing"
	maxOSSURLLength         = 512
)

type SaveResumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveResumeLogic {
	return &SaveResumeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存 app-apis 已上传成功的 OSS 简历元数据；oss_url 用作跨重试幂等键
func (l *SaveResumeLogic) SaveResume(in *pb.SaveResumeReq) (*pb.SaveResumeResp, error) {
	if in.GetUserId() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "用户id不正确")
	}

	// 只保存稳定 HTTPS URL，明确拒绝带签名查询参数的临时下载地址。
	ossUrl := strings.TrimSpace(in.GetOssUrl())
	if !isValidOSSURL(ossUrl) {
		return nil, status.Errorf(codes.InvalidArgument, "oss_url 格式错误(简历文件地址不正确)")
	}

	//常规重试在这里直接返回已有记录，不会重复创建简历
	existing, err := l.svcCtx.ResumesModel.FindOneByOSSURL(l.ctx, ossUrl)
	if err == nil {
		return resumeSaveResponse(existing, in.GetUserId())
	}
	//  判断是不是"记录不存在"
	if !errors.Is(err, model.ErrNotFound) {
		//不是记录不存在，返回错误
		l.Errorf("find resume by OSS URL failed, %v", err)
		return nil, status.Errorf(codes.Internal, "保存简历失败")
	}

	//只有明确未找到相同的URL时才插入
	result, err := l.svcCtx.ResumesModel.Insert(l.ctx, &model.Resumes{
		UserId:     uint64(in.GetUserId()),
		OssUrl:     ossUrl,
		ParseState: defaultResumeParseState,
	})
	if err != nil {
		//判断是不是"唯一键冲突"错误（两个并发请求可能同时通过前置查询，唯一索引冲突后重新查询即可得到同一结果）
		if isDuplicateEntry(err) {
			return l.FindDuplicateResume(ossUrl, in.GetUserId())
		}
		l.Errorf("insert resume failed: %v", err)
		return nil, status.Errorf(codes.Internal, "保存简历失败")
	}
	//直接返回数据库生成的ID, api不需要再次查询
	resumeId, err := result.LastInsertId()
	if err != nil {
		l.Errorf("get insert resume id failed: %v", err)
		return nil, status.Errorf(codes.Internal, "保存简历失败")
	}
	return &pb.SaveResumeResp{
		ResumeId:   resumeId,
		ParseState: defaultResumeParseState,
	}, nil
}

// isValidOSSURL 严格校验 OSS URL 是否合法、安全、永久有效。
func isValidOSSURL(rawUrl string) bool {
	if rawUrl == "" || len(rawUrl) > maxOSSURLLength {
		return false
	}
	//. 解析为绝对 URL（必须以 http:// 或 https:// 开头）
	parsedURL, err := url.ParseRequestURI(rawUrl)
	//左假则不继续后边，返回false
	//逐项检查：必须 HTTPS、有主机、无认证信息、无签名参数、无锚点
	return err == nil &&
		parsedURL.Scheme == "https" &&
		parsedURL.Host != "" &&
		parsedURL.User == nil &&
		parsedURL.RawQuery == "" &&
		parsedURL.Fragment == ""
}

// resumeSaveResponse 保存简历响应校验幂等归属，禁止将其他用户的简历ID返回给当前调用方。
func resumeSaveResponse(resume *model.Resumes, userID int64) (*pb.SaveResumeResp, error) {
	if resume.UserId != uint64(userID) {
		return nil, status.Errorf(codes.AlreadyExists, "该简历不属于当前用户")
	}
	return &pb.SaveResumeResp{
		ResumeId:   int64(resume.Id),
		ParseState: resume.ParseState,
	}, nil

}

// FindDuplicateResume 处理唯一键冲突错误，返回已存在的简历ID
func (l *SaveResumeLogic) FindDuplicateResume(ossURL string, UserId int64) (*pb.SaveResumeResp, error) {
	resume, err := l.svcCtx.ResumesModel.FindOneByOSSURL(l.ctx, ossURL)
	if err != nil {
		l.Errorf("find duplicate resume failed, %v", err)
		return nil, status.Errorf(codes.Internal, "保存简历失败")
	}
	return resumeSaveResponse(resume, UserId)
}

// func (l *SaveResumeLogic) FindDuplicateResume(ossUrl string, userID int64) (*pb.SaveResumeResp, error) {
// 	existing, err := l.svcCtx.ResumesModel.FindOneByOSSURL(l.ctx,ossUrl)
// 	if err == nil {
// 		return resumeSaveResponse(existing, userID)
// 	}
// 	//  判断是不是"记录不存在"
// 	if !errors.Is(err,model.ErrNotFound) {
// 		//不是记录不存在，返回错误
// 		l.Errorf("find resume by OSS URL failed, %v", err)
// 		return nil, status.Errorf(codes.Internal, "保存简历失败")
// 	}
// 	return nil, status.Errorf(codes.AlreadyExists, "该简历已存在")
// }
