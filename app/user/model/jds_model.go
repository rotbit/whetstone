package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// JD 映射 MySQL j
type JD struct {
	ID         uint64         `db:"id"`          // JD 记录主键。
	UserID     uint64         `db:"user_id"`     // 创建 JD 的用户 ID。
	Title      string         `db:"title"`       // 用户填写并去除首尾空白后的标题。
	Content    string         `db:"content"`     // 用户填写并去除首尾空白后的正文。
	ParsedJSON sql.NullString `db:"parsed_json"` // 后续结构化解析结果，未解析时为 NULL。
	CreatedAt  time.Time      `db:"created_at"`  // 数据库生成的创建时间。
}

据库生成的创建时间。
}

// JDsModel
type JDsModel interface {
问能力。
type JDsModel interface {
	Insert(ctx context.Context
	Insert(ctx context.Context, data *JD) (sql.Result, error)
}


type defaultJDsModel struct {
	conn sqlx.SqlConn
}


func NewJDsModel(conn sqlx.SqlConn) JDsModel {
	return &defaultJDsModel{conn: conn}
}

func (m *defaultJDsModel) Insert(ctx context.Context, data *JD) (sql.Result, error) {
	const query = "insert into `jds` (`user_id`,`title`,`content`) values (?, ?, ?)"
	return m.conn.ExecCtx(ctx, query, data.UserID, data.Title, data.Content)
}
