CREATE EXTENSION IF NOT EXISTS vector;
DO $$ BEGIN IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'posts'
        AND column_name = 'embedding'
) THEN
ALTER TABLE posts
ADD COLUMN embedding vector(768);
END IF;
IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'projects'
        AND column_name = 'embedding'
) THEN
ALTER TABLE projects
ADD COLUMN embedding vector(768);
END IF;
END $$;