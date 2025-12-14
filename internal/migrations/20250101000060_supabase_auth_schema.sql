-- Migration: Supabase Authentication Schema
-- Description: Creates auth schema with users and identities tables compatible with Supabase auth
-- Created: 2025-01-01

-- Create auth schema
CREATE SCHEMA IF NOT EXISTS auth;

-- Set up auth schema permissions
REVOKE ALL ON SCHEMA auth FROM public;
GRANT USAGE ON SCHEMA auth TO public;
GRANT ALL ON SCHEMA auth TO postgres;

-- Create users table (main user table)
CREATE TABLE IF NOT EXISTS auth.users (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id uuid REFERENCES auth.instances(id) ON DELETE CASCADE,
    aud text,
    role text,
    email text,
    encrypted_password text,
    email_confirmed_at timestamptz,
    invited_at timestamptz,
    confirmation_token text,
    confirmation_sent_at timestamptz,
    recovery_token text,
    recovery_sent_at timestamptz,
    email_change_token_new text,
    email_change text,
    email_change_sent_at timestamptz,
    last_sign_in_at timestamptz,
    raw_app_meta_data jsonb,
    raw_user_meta_data jsonb,
    is_super_admin boolean,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    phone text,
    phone_confirmed_at timestamptz,
    phone_change text,
    phone_change_token text,
    phone_change_sent_at timestamptz,
    confirmed_at timestamptz,
    email_change_token_current text,
    email_change_confirmed_at timestamptz,
    last_sign_in_ip inet,
    banned_until timestamptz,
    reauthentication_token text,
    reauthentication_sent_at timestamptz,
    is_anonymous boolean,
    deleted_at timestamptz
);

-- Create identities table (for OAuth and other identity providers)
CREATE TABLE IF NOT EXISTS auth.identities (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    identity_data jsonb NOT NULL,
    provider text NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    email text
);

-- Create instances table (for multi-tenant support)
CREATE TABLE IF NOT EXISTS auth.instances (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    uuid uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    raw_base_config jsonb
);

-- Create refresh tokens table
CREATE TABLE IF NOT EXISTS auth.refresh_tokens (
    instance_id uuid REFERENCES auth.instances(id) ON DELETE CASCADE,
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    token text UNIQUE NOT NULL,
    user_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    revoked boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    parent text,
    session_id uuid
);

-- Create schema_migrations table for tracking auth schema migrations
CREATE TABLE IF NOT EXISTS auth.schema_migrations (
    version text PRIMARY KEY,
    dirty boolean NOT NULL DEFAULT false
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS auth_users_instance_id_idx ON auth.users(instance_id);
CREATE INDEX IF NOT EXISTS auth_users_email_idx ON auth.users(email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS auth_identities_user_id_idx ON auth.identities(user_id);
CREATE INDEX IF NOT EXISTS auth_identities_provider_idx ON auth.identities(provider);
CREATE INDEX IF NOT EXISTS auth_refresh_tokens_user_id_idx ON auth.refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS auth_refresh_tokens_token_idx ON auth.refresh_tokens(token);
CREATE INDEX IF NOT EXISTS auth_refresh_tokens_instance_id_idx ON auth.refresh_tokens(instance_id);

-- Create functions for auth operations
CREATE OR REPLACE FUNCTION auth.uid() RETURNS uuid AS $$
    SELECT current_setting('request.jwt.claims', true)::json->>'sub'::uuid;
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to get current user
CREATE OR REPLACE FUNCTION auth.user() RETURNS jsonb AS $$
    SELECT current_setting('request.jwt.claims', true)::jsonb;
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to check if user is authenticated
CREATE OR REPLACE FUNCTION auth.is_authenticated() RETURNS boolean AS $$
    SELECT current_setting('request.jwt.claims', true) IS NOT NULL;
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to get user email
CREATE OR REPLACE FUNCTION auth.email() RETURNS text AS $$
    SELECT current_setting('request.jwt.claims', true)::json->>'email';
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to get user role
CREATE OR REPLACE FUNCTION auth.role() RETURNS text AS $$
    SELECT current_setting('request.jwt.claims', true)::json->>'role';
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to get user metadata
CREATE OR REPLACE FUNCTION auth.get_user_metadata(p_key text) RETURNS text AS $$
BEGIN
    RETURN (
        SELECT raw_app_meta_data->>p_key
        FROM auth.users
        WHERE id = auth.uid()
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to set user metadata
CREATE OR REPLACE FUNCTION auth.set_user_metadata(p_key text, p_value text) RETURNS void AS $$
BEGIN
    UPDATE auth.users
    SET raw_app_meta_data =
        COALESCE(raw_app_meta_data, '{}'::jsonb) || jsonb_build_object(p_key, p_value)
    WHERE id = auth.uid();
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to get all user metadata
CREATE OR REPLACE FUNCTION auth.get_all_user_metadata() RETURNS jsonb AS $$
BEGIN
    RETURN (
        SELECT raw_app_meta_data
        FROM auth.users
        WHERE id = auth.uid()
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to get user by email
CREATE OR REPLACE FUNCTION auth.get_user_by_email(p_email text) RETURNS auth.users AS $$
    SELECT * FROM auth.users WHERE email = p_email LIMIT 1;
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to get user by id
CREATE OR REPLACE FUNCTION auth.get_user_by_id(p_user_id uuid) RETURNS auth.users AS $$
    SELECT * FROM auth.users WHERE id = p_user_id LIMIT 1;
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to create a new user
CREATE OR REPLACE FUNCTION auth.create_user(
    p_email text,
    p_password text,
    p_metadata jsonb DEFAULT NULL
) RETURNS auth.users AS $$
DECLARE
    user_record auth.users;
    hashed_password text;
BEGIN
    -- Hash the password (in a real implementation, use a proper hashing function)
    -- For now, we'll just store it as-is (this would be replaced with proper hashing)
    hashed_password := p_password;

    -- Insert the new user
    INSERT INTO auth.users (
        email,
        encrypted_password,
        raw_app_meta_data,
        raw_user_meta_data,
        created_at,
        updated_at
    ) VALUES (
        p_email,
        hashed_password,
        p_metadata,
        '{}'::jsonb,
        now(),
        now()
    ) RETURNING * INTO user_record;

    RETURN user_record;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to verify user password
CREATE OR REPLACE FUNCTION auth.verify_password(
    p_user_id uuid,
    p_password text
) RETURNS boolean AS $$
DECLARE
    stored_password text;
BEGIN
    SELECT encrypted_password INTO stored_password
    FROM auth.users
    WHERE id = p_user_id;

    -- In a real implementation, this would use proper password verification
    -- For now, we'll do a simple comparison (this would be replaced)
    RETURN stored_password = p_password;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to update user password
CREATE OR REPLACE FUNCTION auth.update_password(
    p_user_id uuid,
    p_new_password text
) RETURNS void AS $$
BEGIN
    UPDATE auth.users
    SET
        encrypted_password = p_new_password, -- In real implementation, hash this
        updated_at = now()
    WHERE id = p_user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to confirm user email
CREATE OR REPLACE FUNCTION auth.confirm_email(p_user_id uuid) RETURNS void AS $$
BEGIN
    UPDATE auth.users
    SET
        email_confirmed_at = now(),
        confirmed_at = now(),
        updated_at = now()
    WHERE id = p_user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to get user organization ID
CREATE OR REPLACE FUNCTION auth.get_user_organization_id(p_user_id uuid) RETURNS uuid AS $$
BEGIN
    RETURN (
        SELECT (raw_app_meta_data->>'organization_id')::uuid
        FROM auth.users
        WHERE id = p_user_id
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to set user organization ID
CREATE OR REPLACE FUNCTION auth.set_user_organization_id(
    p_user_id uuid,
    p_organization_id uuid
) RETURNS void AS $$
BEGIN
    UPDATE auth.users
    SET raw_app_meta_data =
        COALESCE(raw_app_meta_data, '{}'::jsonb) ||
        jsonb_build_object('organization_id', p_organization_id)
    WHERE id = p_user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to check if user belongs to organization
CREATE OR REPLACE FUNCTION auth.user_belongs_to_organization(
    p_user_id uuid,
    p_organization_id uuid
) RETURNS boolean AS $$
BEGIN
    RETURN (
        SELECT (raw_app_meta_data->>'organization_id')::uuid = p_organization_id
        FROM auth.users
        WHERE id = p_user_id
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to create an identity (for OAuth providers)
CREATE OR REPLACE FUNCTION auth.create_identity(
    p_user_id uuid,
    p_provider text,
    p_identity_data jsonb
) RETURNS auth.identities AS $$
DECLARE
    identity_record auth.identities;
BEGIN
    INSERT INTO auth.identities (
        user_id,
        provider,
        identity_data,
        created_at,
        updated_at
    ) VALUES (
        p_user_id,
        p_provider,
        p_identity_data,
        now(),
        now()
    ) RETURNING * INTO identity_record;

    RETURN identity_record;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to get identities by user
CREATE OR REPLACE FUNCTION auth.get_user_identities(p_user_id uuid) RETURNS SETOF auth.identities AS $$
    SELECT * FROM auth.identities WHERE user_id = p_user_id;
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to get identity by provider and user
CREATE OR REPLACE FUNCTION auth.get_identity_by_provider(
    p_user_id uuid,
    p_provider text
) RETURNS auth.identities AS $$
    SELECT * FROM auth.identities
    WHERE user_id = p_user_id AND provider = p_provider LIMIT 1;
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to update identity data
CREATE OR REPLACE FUNCTION auth.update_identity(
    p_identity_id uuid,
    p_identity_data jsonb
) RETURNS void AS $$
BEGIN
    UPDATE auth.identities
    SET
        identity_data = p_identity_data,
        updated_at = now()
    WHERE id = p_identity_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to delete identity
CREATE OR REPLACE FUNCTION auth.delete_identity(p_identity_id uuid) RETURNS void AS $$
BEGIN
    DELETE FROM auth.identities WHERE id = p_identity_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to create a refresh token
CREATE OR REPLACE FUNCTION auth.create_refresh_token(
    p_user_id uuid,
    p_token text,
    p_session_id uuid
) RETURNS auth.refresh_tokens AS $$
DECLARE
    token_record auth.refresh_tokens;
BEGIN
    INSERT INTO auth.refresh_tokens (
        user_id,
        token,
        session_id,
        created_at,
        updated_at
    ) VALUES (
        p_user_id,
        p_token,
        p_session_id,
        now(),
        now()
    ) RETURNING * INTO token_record;

    RETURN token_record;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to revoke refresh token
CREATE OR REPLACE FUNCTION auth.revoke_refresh_token(p_token text) RETURNS void AS $$
BEGIN
    UPDATE auth.refresh_tokens
    SET
        revoked = true,
        updated_at = now()
    WHERE token = p_token;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to get refresh token
CREATE OR REPLACE FUNCTION auth.get_refresh_token(p_token text) RETURNS auth.refresh_tokens AS $$
    SELECT * FROM auth.refresh_tokens WHERE token = p_token LIMIT 1;
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to validate refresh token
CREATE OR REPLACE FUNCTION auth.validate_refresh_token(p_token text) RETURNS boolean AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM auth.refresh_tokens
        WHERE token = p_token AND revoked = false
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to get user by refresh token
CREATE OR REPLACE FUNCTION auth.get_user_by_refresh_token(p_token text) RETURNS auth.users AS $$
BEGIN
    RETURN (
        SELECT u.*
        FROM auth.users u
        JOIN auth.refresh_tokens rt ON u.id = rt.user_id
        WHERE rt.token = p_token AND rt.revoked = false
        LIMIT 1
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to create auth instance
CREATE OR REPLACE FUNCTION auth.create_instance() RETURNS auth.instances AS $$
DECLARE
    instance_record auth.instances;
BEGIN
    INSERT INTO auth.instances (raw_base_config)
    VALUES ('{}'::jsonb)
    RETURNING * INTO instance_record;

    RETURN instance_record;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to get current instance
CREATE OR REPLACE FUNCTION auth.get_current_instance() RETURNS auth.instances AS $$
    SELECT * FROM auth.instances LIMIT 1;
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to get instance by ID
CREATE OR REPLACE FUNCTION auth.get_instance_by_id(p_instance_id uuid) RETURNS auth.instances AS $$
    SELECT * FROM auth.instances WHERE id = p_instance_id LIMIT 1;
$$ LANGUAGE sql SECURITY DEFINER;

-- Create function to update instance configuration
CREATE OR REPLACE FUNCTION auth.update_instance_config(
    p_instance_id uuid,
    p_config jsonb
) RETURNS void AS $$
BEGIN
    UPDATE auth.instances
    SET raw_base_config = p_config
    WHERE id = p_instance_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to initialize auth schema
CREATE OR REPLACE FUNCTION auth.initialize() RETURNS void AS $$
BEGIN
    -- Create default instance if none exists
    PERFORM auth.create_instance();

    -- Set up default permissions
    EXECUTE 'GRANT SELECT ON ALL TABLES IN SCHEMA auth TO public';
    EXECUTE 'GRANT USAGE ON ALL SEQUENCES IN SCHEMA auth TO public';
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create triggers for automatic timestamp updates
CREATE OR REPLACE FUNCTION auth.update_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER auth_users_updated_at_trigger
BEFORE UPDATE ON auth.users
FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at();

CREATE TRIGGER auth_identities_updated_at_trigger
BEFORE UPDATE ON auth.identities
FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at();

CREATE TRIGGER auth_refresh_tokens_updated_at_trigger
BEFORE UPDATE ON auth.refresh_tokens
FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at();

-- Create initial instance
INSERT INTO auth.instances (raw_base_config)
SELECT '{}'::jsonb
WHERE NOT EXISTS (SELECT 1 FROM auth.instances);

-- Insert initial schema migration record
INSERT INTO auth.schema_migrations (version, dirty)
VALUES ('20250101000060', false)
ON CONFLICT (version) DO NOTHING;
