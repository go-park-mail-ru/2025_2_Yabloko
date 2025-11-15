-- Таблица тикетов поддержки
CREATE TABLE IF NOT EXISTS support_ticket (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES account(id) ON DELETE SET NULL,
    guest_id TEXT NULL,
    user_name TEXT NOT NULL CHECK (length(user_name) <= 100),
    user_email TEXT NOT NULL CHECK (length(user_email) <= 100),
    category TEXT NOT NULL CHECK (category IN ('bug', 'feature', 'complaint')),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'closed')),
    priority TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    title TEXT NOT NULL CHECK (length(title) <= 200),
    description TEXT NOT NULL CHECK (length(description) <= 2000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    
    CONSTRAINT either_user_or_guest CHECK (
        (user_id IS NOT NULL AND guest_id IS NULL) OR 
        (user_id IS NULL AND guest_id IS NOT NULL)
    )
);

-- Таблица сообщений в тикетах
CREATE TABLE IF NOT EXISTS support_message (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_ticket(id) ON DELETE CASCADE,
    user_id UUID NULL REFERENCES account(id) ON DELETE SET NULL,
    guest_id TEXT NULL,
    user_role TEXT NOT NULL CHECK (user_role IN ('user', 'guest', 'admin', 'support')),
    content TEXT NOT NULL CHECK (length(content) <= 2000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    
    CONSTRAINT either_user_or_guest CHECK (
        (user_id IS NOT NULL AND guest_id IS NULL) OR 
        (user_id IS NULL AND guest_id IS NOT NULL)
    )
);

-- Таблица рейтингов тикетов
CREATE TABLE IF NOT EXISTS support_rating (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL UNIQUE REFERENCES support_ticket(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT CHECK (length(comment) <= 500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_support_ticket_user_id ON support_ticket(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_support_ticket_guest_id ON support_ticket(guest_id) WHERE guest_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_support_ticket_status ON support_ticket(status);
CREATE INDEX IF NOT EXISTS idx_support_ticket_category ON support_ticket(category);
CREATE INDEX IF NOT EXISTS idx_support_ticket_created_at ON support_ticket(created_at);
CREATE INDEX IF NOT EXISTS idx_support_message_ticket_id ON support_message(ticket_id);
CREATE INDEX IF NOT EXISTS idx_support_message_created_at ON support_message(created_at);
CREATE INDEX IF NOT EXISTS idx_support_rating_ticket_id ON support_rating(ticket_id);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = current_timestamp;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER trg_update_support_ticket_updated_at
    BEFORE UPDATE ON support_ticket
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
