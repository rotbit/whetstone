package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ResumesModel = (*customResumesModel)(nil)

type (
	// ResumesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customResumesModel.
	ResumesModel interface {
		resumesModel
		withSession(session sqlx.Session) ResumesModel
		// FindOneByOSSURL 按 OSS 稳定 URL 查简历；SaveResume 用它做幂等去重。
		// 找不到时返回 model.ErrNotFound（即 sqlx.ErrNotFound），由调用方区分。
		FindOneByOSSURL(ctx context.Context, ossUrl string) (*Resumes, error)
		// FindLatestByUserID 返回 user_id 下最新一条简历（按 created_at desc + id desc 保证稳定）。
		// 典型用法：GetResume RPC、用户侧查看当前简历。
		FindLatestByUserID(ctx context.Context, userID uint64) (*Resumes, error)
	}

	customResumesModel struct {
		*defaultResumesModel
	}
)

// NewResumesModel returns a model for the database table.
func NewResumesModel(conn sqlx.SqlConn) ResumesModel {
	return &customResumesModel{
		defaultResumesModel: newResumesModel(conn),
	}
}

func (m *customResumesModel) withSession(session sqlx.Session) ResumesModel {
	return NewResumesModel(sqlx.NewSqlConnFromSession(session))
}

// FindOneByOSSURL 按 oss_url 唯一键查询。对应 SQL：
//
//	SELECT id, user_id, oss_url, parsed_json, parse_state, created_at
//	  FROM resumes
//	 WHERE oss_url = ? LIMIT 1
//
// 建议在 resumes 表上建唯一索引 UNIQUE KEY uk_oss_url (oss_url)
// 或更严格的 UNIQUE KEY uk_user_oss (user_id, oss_url) 防止跨用户串数据。
func (m *customResumesModel) FindOneByOSSURL(ctx context.Context, ossUrl string) (*Resumes, error) {
	var resp Resumes
	query := fmt.Sprintf("select %s from %s where `oss_url` = ? limit 1", resumesRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, ossUrl)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindLatestByUserID 查询 user_id 下按创建时间倒序的最新一条简历。
//
//	SELECT id, user_id, oss_url, parsed_json, parse_state, created_at
//	  FROM resumes
//	 WHERE user_id = ?
//	 ORDER BY created_at DESC, id DESC
//	 LIMIT 1
//
// 二级排序 `id desc` 用于解决同一秒内上传多份简历时的结果不稳定问题；
// 表上已有 `KEY idx_user (user_id)`，在小结果集中内存排序不构成性能问题。
func (m *customResumesModel) FindLatestByUserID(ctx context.Context, userID uint64) (*Resumes, error) {
	var resp Resumes
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `created_at` desc, `id` desc limit 1", resumesRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
