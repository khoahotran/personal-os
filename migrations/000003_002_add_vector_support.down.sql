ALTER TABLE posts
DROP COLUMN IF EXISTS embedding;

ALTER TABLE projects
DROP COLUMN IF EXISTS embedding;
-- No DROP EXTENSION statement to avoid removing the extension if other parts of the database are using it.