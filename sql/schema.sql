-- whetstone 核心表（docs/技术方案.md §7）
-- 字符集 utf8mb4；金额单位：分；时间戳统一 datetime

CREATE DATABASE IF NOT EXISTS whetstone DEFAULT CHARACTER SET utf8mb4;
USE whetstone;

CREATE TABLE IF NOT EXISTS users (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    phone       VARCHAR(20)  NOT NULL,
    password    VARCHAR(128) NOT NULL DEFAULT '',
    plan        VARCHAR(20)  NOT NULL DEFAULT 'free', -- free | pack | monthly
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_phone (phone)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS resumes (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id     BIGINT UNSIGNED NOT NULL,
    oss_url     VARCHAR(512) NOT NULL DEFAULT '',
    parsed_json JSON NULL,                              -- LLM 结构化：项目/技能/经历
    parse_state VARCHAR(20) NOT NULL DEFAULT 'parsing', -- parsing | done | failed
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_user (user_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS jds (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id     BIGINT UNSIGNED NOT NULL,
    title       VARCHAR(128) NOT NULL,
    content     TEXT NOT NULL,
    parsed_json JSON NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_user (user_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS interview_sessions (
    id          VARCHAR(64) PRIMARY KEY,               -- uuid
    user_id     BIGINT UNSIGNED NOT NULL,
    position    VARCHAR(64) NOT NULL,
    mode        VARCHAR(10) NOT NULL DEFAULT 'text',   -- text | voice
    resume_id   BIGINT UNSIGNED NULL,
    jd_id       BIGINT UNSIGNED NULL,
    state       VARCHAR(20) NOT NULL DEFAULT 'ongoing',-- ongoing | finished | reported
    started_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at    DATETIME NULL,
    KEY idx_user_time (user_id, started_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS qa_records (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_id     VARCHAR(64) NOT NULL,
    seq            INT NOT NULL,
    question       TEXT NOT NULL,
    answer         TEXT NULL,
    followup_depth TINYINT NOT NULL DEFAULT 0,
    score          TINYINT NULL,                       -- 0-100，report-worker 异步回填
    feedback       TEXT NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_session_seq (session_id, seq)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS reports (
    session_id  VARCHAR(64) PRIMARY KEY,               -- 一场一报告，天然幂等
    radar_json  JSON NULL,                             -- 能力雷达
    summary     TEXT NULL,
    suggestions TEXT NULL,
    state       VARCHAR(20) NOT NULL DEFAULT 'generating', -- generating | done | failed
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS credits (
    user_id    BIGINT UNSIGNED PRIMARY KEY,
    balance    INT NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- 次数流水：session_id 唯一键实现幂等扣减（面试考点 §12）
CREATE TABLE IF NOT EXISTS credit_logs (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id    BIGINT UNSIGNED NOT NULL,
    session_id VARCHAR(64) NOT NULL,
    delta      INT NOT NULL,                           -- 负数扣减 / 正数回补或充值
    reason     VARCHAR(32) NOT NULL,                   -- deduct | refund | purchase | gift
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_session_reason (session_id, reason),
    KEY idx_user (user_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS orders (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    out_trade_no VARCHAR(64) NOT NULL,
    user_id      BIGINT UNSIGNED NOT NULL,
    sku          VARCHAR(32) NOT NULL,                 -- pack_3 | monthly
    amount       INT NOT NULL,                         -- 分
    state        VARCHAR(20) NOT NULL DEFAULT 'created', -- created | paying | paid | closed
    paid_at      DATETIME NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_trade_no (out_trade_no),
    KEY idx_user (user_id)
) ENGINE=InnoDB;

-- 题库（question-rpc；向量存 Qdrant，MySQL 存原文与标签）
CREATE TABLE IF NOT EXISTS questions (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    position   VARCHAR(64) NOT NULL,                   -- go-backend | frontend | ...
    category   VARCHAR(20) NOT NULL,                   -- basic | project | scenario | algorithm
    level      TINYINT NOT NULL DEFAULT 1,             -- 1-5
    content    TEXT NOT NULL,
    tags       VARCHAR(255) NOT NULL DEFAULT '',       -- 逗号分隔技能标签
    reference  TEXT NULL,                              -- 参考答案要点
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_position_category (position, category)
) ENGINE=InnoDB;
