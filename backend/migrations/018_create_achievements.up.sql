-- Migration 018: Create achievements system
-- Created: 2024

-- Create achievements table for system-wide achievements
CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    condition_json JSONB NOT NULL,
    points INT DEFAULT 0,
    is_hidden BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create index for lookup by code
CREATE INDEX IF NOT EXISTS idx_achievements_code ON achievements(code);
CREATE INDEX IF NOT EXISTS idx_achievements_hidden ON achievements(is_hidden);

-- Create user_achievements table to track earned achievements
CREATE TABLE IF NOT EXISTS user_achievements (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    obtained_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata JSONB DEFAULT '{}',
    PRIMARY KEY (user_id, achievement_id)
);

-- Create index for user achievement lookups
CREATE INDEX IF NOT EXISTS idx_user_achievements_user ON user_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_user_achievements_achievement ON user_achievements(achievement_id);

-- Seed default achievements
INSERT INTO achievements (code, title, description, icon, condition_json, points, is_hidden) VALUES
    (
        'first_note',
        'Первые Искры',
        'Создана первая заметка во вселенной',
        'star',
        '{"type": "count", "entity": "note", "action": "create", "threshold": 1}',
        10,
        false
    ),
    (
        'note_master_10',
        'Архитектор Галактики',
        'Создано 10 заметок',
        'planet',
        '{"type": "count", "entity": "note", "action": "create", "threshold": 10}',
        50,
        false
    ),
    (
        'note_master_50',
        'Космический Коллекционер',
        'Создано 50 заметок',
        'galaxy',
        '{"type": "count", "entity": "note", "action": "create", "threshold": 50}',
        200,
        false
    ),
    (
        'link_maker',
        'Ткачь Гравитации',
        'Создана первая связь между объектами',
        'link',
        '{"type": "count", "entity": "link", "action": "create", "threshold": 1}',
        15,
        false
    ),
    (
        'link_master',
        'Повелитель Пространства',
        'Создано 20 связей',
        'network',
        '{"type": "count", "entity": "link", "action": "create", "threshold": 20}',
        100,
        false
    ),
    (
        'galaxy_builder',
        'Строитель Млечного Пути',
        'Создано 5 заметок типа "galaxy"',
        'sparkles',
        '{"type": "count", "entity": "note", "action": "create", "filter": {"type": "galaxy"}, "threshold": 5}',
        75,
        false
    ),
    (
        'star_collector',
        'Собиратель Звёзд',
        'Создано 10 заметок типа "star"',
        'sun',
        '{"type": "count", "entity": "note", "action": "create", "filter": {"type": "star"}, "threshold": 10}',
        60,
        false
    ),
    (
        'seven_day_streak',
        'Непрерывный Путник',
        'Входил в систему 7 дней подряд',
        'flame',
        '{"type": "streak", "action": "login", "threshold": 7}',
        150,
        false
    ),
    (
        'explorer',
        'Искатель Тайн',
        'Использовал поиск 50 раз',
        'search',
        '{"type": "count", "entity": "search", "action": "execute", "threshold": 50}',
        80,
        true
    ),
    (
        'sharer',
        'Даритель Света',
        'Поделился заметкой с другим пользователем',
        'share',
        '{"type": "count", "entity": "share", "action": "create", "threshold": 1}',
        25,
        false
    )
ON CONFLICT (code) DO NOTHING;

COMMIT;
