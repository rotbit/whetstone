package userlogic

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/user/rpc/pb"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultPlan = "free"

var phonePattern = regexp.MustCompile(`^\+?[0-9]{6,20}$`)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *pb.RegisterReq) (*pb.RegisterResp, error) {
	phone := strings.TrimSpace(in.GetPhone())
	if !phonePattern.MatchString(phone) {
		return nil, status.Error(codes.InvalidArgument, "手机号格式不正确")
	}

	password := in.GetPassword()
	if len(password) < 8 || len(password) > 72 {
		return nil, status.Error(codes.InvalidArgument, "密码长度必须为 8 到 72 个字符")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		l.Errorf("generate password hash failed: %v", err)
		return nil, status.Error(codes.Internal, "注册失败")
	}

	result, err := l.svcCtx.UsersModel.Insert(l.ctx, &model.Users{
		Phone:    phone,
		Password: string(passwordHash),
		Plan:     defaultPlan,
	})
	if err != nil {
		if isDuplicateEntry(err) {
			return nil, status.Error(codes.AlreadyExists, "手机号已注册")
		}

		l.Errorf("insert user failed: %v", err)
		return nil, status.Error(codes.Internal, "注册失败")
	}

	userID, err := result.LastInsertId()
	if err != nil {
		l.Errorf("get inserted user id failed: %v", err)
		return nil, status.Error(codes.Internal, "注册失败")
	}

	return &pb.RegisterResp{
		UserId: userID,
		Phone:  phone,
		Plan:   defaultPlan,
	}, nil
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
