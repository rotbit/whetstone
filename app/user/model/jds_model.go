package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// JD 映射 MySQL jds 表的一行目标岗位描述。
type JD struct {
	ID         uint64         `db:"id"`          // JD 记录主键。
	UserID     uint64         `db:"user_id"`     // 创建 JD 的用户 ID。
	Title      string         `db:"title"`       // 用户填写并去除首尾空白后的标题。
	Content    string         `db:"content"`     // 用户填写并去除首尾空白后的正文。
	ParsedJSON sql.NullString `db:"parsed_json"` // 后续结构化解析结果，未解析时为 NULL。
	CreatedAt  time.Time      `db:"created_at"`  // 数据库生成的创建时间。
}

// JDsModel 定义 user-rpc 当前需要的最小 JD 数据访问能力。
type JDsModel interface {
	Insert(ctx context.Context, data *JD) (sql.Result, error)
}

// defaultJDsModel 使用 go-zero sqlx 执行 MySQL 写入。
type defaultJDsModel struct {
	conn sqlx.SqlConn
}

// NewJDsModel 创建 JD 数据访问对象，并复用调用方提供的数据库连接。
func NewJDsModel(conn sqlx.SqlConn) JDsModel {
	return &defaultJDsModel{conn: conn}
}

// Insert 保存原始 JD；parsed_json 和 created_at 使用数据库默认值。
func (m *defaultJDsModel) Insert(ctx context.Context, data *JD) (sql.Result, error) {
	const query = "insert into `jds` (`user_id`,`title`,`content`) values (?, ?, ?)"
	return m.conn.ExecCtx(ctx, query, data.UserID, data.Title, data.Content)
}
