-- Apply this migration only after the duplicate check returns no rows.
SELECT oss_url, COUNT(*) AS duplicate_count
FROM resumes
GROUP BY oss_url
HAVING COUNT(*) > 1;

ALTER TABLE resumes
    ADD UNIQUE KEY uk_oss_url (oss_url);
