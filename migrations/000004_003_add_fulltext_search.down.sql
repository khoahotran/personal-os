DROP INDEX IF EXISTS projects_ts_idx;
DROP TRIGGER IF EXISTS tsvectorupdate_projects ON projects;
DROP FUNCTION IF EXISTS projects_tsvector_trigger();
ALTER TABLE projects DROP COLUMN IF EXISTS ts;

DROP INDEX IF EXISTS posts_ts_idx;
DROP TRIGGER IF EXISTS tsvectorupdate_posts ON posts;
DROP FUNCTION IF EXISTS posts_tsvector_trigger();
ALTER TABLE posts DROP COLUMN IF EXISTS ts;