DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'apple_user') THEN
        CREATE USER apple_user WITH PASSWORD 'postgres';
        RAISE NOTICE 'User apple_user created';
    ELSE
        RAISE NOTICE 'User apple_user already exists';
    END IF;
END
$$;