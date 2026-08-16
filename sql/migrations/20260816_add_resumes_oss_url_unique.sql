-- oss_url 是 SaveResume 的幂等键；增加唯一索引前，必须先确认历史数据没有重复 URL。
-- 如果下方查询返回任何记录，请先人工合并或清理重复数据，不要继续执行 ALTER TABLE。
SELECT oss_url, COUNT(*) AS duplicate_count
FROM resumes
GROUP BY oss_url
HAVING COUNT(*) > 1;

-- 唯一索引同时覆盖 API 重试和多个实例并发写入，最终只会保留一条简历记录。
ALTER TABLE resumes
    ADD UNIQUE KEY uk_oss_url (oss_url);
