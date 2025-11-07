--Posts
ALTER TABLE posts
ADD COLUMN ts tsvector;
CREATE OR REPLACE FUNCTION posts_tsvector_trigger() RETURNS trigger AS $$ BEGIN NEW.ts := setweight(
        to_tsvector('simple', COALESCE(NEW.title, '')),
        'A'
    ) || setweight(
        to_tsvector('simple', COALESCE(NEW.content_markdown, '')),
        'B'
    );
RETURN NEW;
END $$ LANGUAGE plpgsql;
CREATE TRIGGER tsvectorupdate_posts BEFORE
INSERT
    OR
UPDATE ON posts FOR EACH ROW EXECUTE PROCEDURE posts_tsvector_trigger();
CREATE INDEX posts_ts_idx ON posts USING GIN(ts);
-- Projects
ALTER TABLE projects
ADD COLUMN ts tsvector;
CREATE OR REPLACE FUNCTION projects_tsvector_trigger() RETURNS trigger AS $$ BEGIN NEW.ts := setweight(
        to_tsvector('simple', COALESCE(NEW.title, '')),
        'A'
    ) || setweight(
        to_tsvector('simple', COALESCE(NEW.description, '')),
        'B'
    );
RETURN NEW;
END $$ LANGUAGE plpgsql;
CREATE TRIGGER tsvectorupdate_projects BEFORE
INSERT
    OR
UPDATE ON projects FOR EACH ROW EXECUTE PROCEDURE projects_tsvector_trigger();
CREATE INDEX projects_ts_idx ON projects USING GIN(ts);