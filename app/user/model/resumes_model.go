package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// resumeFields 集中维护查询列，确保不同查询返回的 Resume 字段保持一致。
const resumeFields = "`id`,`user_id`,`oss_url`,`parsed_json`,`parse_state`,`created_at`"

// Resume 映射 MySQL resumes 表的一行简历元数据。
// 文件本体保存在 OSS；MySQL 只保存稳定 URL、解析状态和后续解析结果。
type Resume struct {
	ID         uint64         `db:"id"`          // 简历记录主键。
	UserID     uint64         `db:"user_id"`     // 上传简历的用户 ID。
	OSSURL     string         `db:"oss_url"`     // 不带临时签名的稳定 OSS URL，同时作为幂等键。
	ParsedJSON sql.NullString `db:"parsed_json"` // 尚未解析时为 NULL，不能用普通 string 区分。
	ParseState string         `db:"parse_state"` // parsing、done 或 failed。
	CreatedAt  time.Time      `db:"created_at"`  // 数据库生成的创建时间。
}

// ResumesModel 定义 user-rpc 当前需要的最小简历数据访问能力。
type ResumesModel interface {
	Insert(ctx context.Context, data *Resume) (sql.Result, error)
	FindLatestByUserID(ctx context.Context, userID uint64) (*Resume, error)
	FindOneByOSSURL(ctx context.Context, ossURL string) (*Resume, error)
}

// defaultResumesModel 使用 go-zero sqlx 执行 MySQL 查询。
type defaultResumesModel struct {
	conn sqlx.SqlConn
}

// NewResumesModel 创建简历数据访问对象，并复用调用方提供的数据库连接。
func NewResumesModel(conn sqlx.SqlConn) ResumesModel {
	return &defaultResumesModel{conn: conn}
}

// Insert 新建简历元数据；parsed_json 和 created_at 使用数据库默认值。
func (m *defaultResumesModel) Insert(ctx context.Context, data *Resume) (sql.Result, error) {
	const query = "insert into `resumes` (`user_id`,`oss_url`,`parse_state`) values (?, ?, ?)"
	return m.conn.ExecCtx(ctx, query, data.UserID, data.OSSURL, data.ParseState)
}

// FindLatestByUserID 按自增主键倒序返回用户最后上传的一份简历。
// InnoDB 二级索引包含主键值，因此 idx_user 可支持该用户范围内的最新记录查询。
func (m *defaultResumesModel) FindLatestByUserID(ctx context.Context, userID uint64) (*Resume, error) {
	query := "select " + resumeFields + " from `resumes` where `user_id` = ? order by `id` desc limit 1"
	return m.findOne(ctx, query, userID)
}

// FindOneByOSSURL 根据稳定 URL 查询记录，用于 SaveResume 的应用层幂等检查。
func (m *defaultResumesModel) FindOneByOSSURL(ctx context.Context, ossURL string) (*Resume, error) {
	query := "select " + resumeFields + " from `resumes` where `oss_url` = ? limit 1"
	return m.findOne(ctx, query, ossURL)
}

// findOne 复用单行扫描逻辑，并把 sqlx 的未找到错误转换为 model 包统一错误。
func (m *defaultResumesModel) findOne(ctx context.Context, query string, arg any) (*Resume, error) {
	var resume Resume
	err := m.conn.QueryRowCtx(ctx, &resume, query, arg)
	if err == nil {
		return &resume, nil
	}
	if err == sqlx.ErrNotFound {
		return nil, ErrNotFound
	}
	return nil, err
}
