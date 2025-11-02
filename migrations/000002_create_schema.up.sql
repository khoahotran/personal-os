-- Function update_updated_at_column
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- 1. USERS
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    profile_settings JSONB DEFAULT '{}'
);
-- 2. PROFILES
CREATE TABLE IF NOT EXISTS profiles (
    owner_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    career_timeline JSONB DEFAULT '[]',
    theme_settings JSONB DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
DROP TRIGGER IF EXISTS update_profiles_updated_at ON profiles;
CREATE TRIGGER update_profiles_updated_at BEFORE
UPDATE ON profiles FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
-- 3. PROJECTS
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    stack TEXT [],
    repository_url VARCHAR(255),
    live_url VARCHAR(255),
    media JSONB DEFAULT '[]',
    is_public BOOLEAN DEFAULT false NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_slug ON projects(slug);
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
CREATE TRIGGER update_projects_updated_at BEFORE
UPDATE ON projects FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
-- 4. POSTS
CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    content_markdown TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    og_image_url VARCHAR(255),
    thumbnail_url VARCHAR(255),
    metadata JSONB DEFAULT '{}'::jsonb,
    version_history JSONB DEFAULT '[]',
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT posts_status_check CHECK (
        status IN ('draft', 'private', 'public', 'pending')
    )
);
CREATE INDEX IF NOT EXISTS idx_posts_owner_id_status ON posts(owner_id, status);
CREATE INDEX IF NOT EXISTS idx_posts_metadata ON posts USING gin (metadata);
CREATE INDEX IF NOT EXISTS idx_posts_slug ON posts(slug);
DROP TRIGGER IF EXISTS update_posts_updated_at ON posts;
CREATE TRIGGER update_posts_updated_at BEFORE
UPDATE ON posts FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
-- 5. POST_VERSIONS
CREATE TABLE IF NOT EXISTS post_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    content_diff TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_post_versions_post_id ON post_versions(post_id);
-- 6. MEDIA
CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    url TEXT NOT NULL,
    thumbnail_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    metadata JSONB DEFAULT '{}',
    is_public BOOLEAN DEFAULT false NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER update_media_updated_at BEFORE
UPDATE ON media FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
-- 7. HOBBY_ITEMS
CREATE TABLE IF NOT EXISTS hobby_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status VARCHAR(50),
    rating SMALLINT,
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    is_public BOOLEAN DEFAULT false NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hobby_items_owner_id_category ON hobby_items(owner_id, category);
DROP TRIGGER IF EXISTS update_hobby_items_updated_at ON hobby_items;
CREATE TRIGGER update_hobby_items_updated_at BEFORE
UPDATE ON hobby_items FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
-- 8. TAGS
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL
);
-- 9. TAG_RELATIONS
CREATE TABLE IF NOT EXISTS tag_relations (
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    resource_id UUID NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    PRIMARY KEY (tag_id, resource_id, resource_type)
);
CREATE INDEX IF NOT EXISTS idx_tag_relations_resource ON tag_relations(resource_id, resource_type);