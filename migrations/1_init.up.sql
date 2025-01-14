CREATE TABLE IF NOT EXISTS public.files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL, 
    data BYTEA NOT NULL,
    create_time timestamp default current_timestamp,
    update_time timestamp default current_timestamp
);