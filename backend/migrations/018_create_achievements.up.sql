-- Migration 018: Create achievements system
-- Updated: 2026-06-01 — aligned with AchievementModel (name_ru/en, description_ru/en, icon_emoji, category, hidden)

BEGIN;

-- Create achievements table (matches achievement_model.go)
CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) NOT NULL UNIQUE,
    name_ru TEXT,
    name_en TEXT,
    description_ru TEXT,
    description_en TEXT,
    icon_emoji VARCHAR(8),
    category VARCHAR(50),
    condition_json JSONB NOT NULL DEFAULT '{}',
    points INTEGER NOT NULL DEFAULT 0,
    hidden BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_achievements_code ON achievements(code);
CREATE INDEX IF NOT EXISTS idx_achievements_hidden ON achievements(hidden);

-- Create user_achievements table (matches user_achievement_model.go)
CREATE TABLE IF NOT EXISTS user_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    notification_seen BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, achievement_id)
);

CREATE INDEX IF NOT EXISTS idx_user_achievements_user ON user_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_user_achievements_achievement ON user_achievements(achievement_id);

-- Seed default achievements
INSERT INTO achievements (code, name_ru, name_en, description_ru, description_en, icon_emoji, category, condition_json, points, hidden) VALUES
    (
        'first_note',
        'Первые Искры',
        'First Sparks',
        'Создана первая заметка во вселенной',
        'Created your first note in the universe',
        '⭐',
        'creation',
        '{"type": "count", "entity": "note", "action": "create", "threshold": 1}',
        10,
        false
    ),
    (
        'note_master_10',
        'Архитектор Галактики',
        'Galaxy Architect',
        'Создано 10 заметок',
        'Created 10 notes',
        '🪐',
        'creation',
        '{"type": "count", "entity": "note", "action": "create", "threshold": 10}',
        50,
        false
    ),
    (
        'note_master_50',
        'Космический Коллекционер',
        'Cosmic Collector',
        'Создано 50 заметок',
        'Created 50 notes',
        '🌌',
        'creation',
        '{"type": "count", "entity": "note", "action": "create", "threshold": 50}',
        200,
        false
    ),
    (
        'link_maker',
        'Ткачь Гравитации',
        'Gravity Weaver',
        'Создана первая связь между объектами',
        'Created your first link between objects',
        '🔗',
        'connection',
        '{"type": "count", "entity": "link", "action": "create", "threshold": 1}',
        15,
        false
    ),
    (
        'link_master',
        'Повелитель Пространства',
        'Space Lord',
        'Создано 20 связей',
        'Created 20 links',
        '🕸️',
        'connection',
        '{"type": "count", "entity": "link", "action": "create", "threshold": 20}',
        100,
        false
    ),
    (
        'galaxy_builder',
        'Строитель Млечного Пути',
        'Milky Way Builder',
        'Создано 5 заметок типа "galaxy"',
        'Created 5 notes of type "galaxy"',
        '✨',
        'creation',
        '{"type": "count", "entity": "note", "action": "create", "filter": {"type": "galaxy"}, "threshold": 5}',
        75,
        false
    ),
    (
        'star_collector',
        'Собиратель Звёзд',
        'Star Collector',
        'Создано 10 заметок типа "star"',
        'Created 10 notes of type "star"',
        '☀️',
        'creation',
        '{"type": "count", "entity": "note", "action": "create", "filter": {"type": "star"}, "threshold": 10}',
        60,
        false
    ),
    (
        'seven_day_streak',
        'Непрерывный Путник',
        'Continuous Traveler',
        'Входил в систему 7 дней подряд',
        'Logged in for 7 consecutive days',
        '🔥',
        'streak',
        '{"type": "streak", "action": "login", "threshold": 7}',
        150,
        false
    ),
    (
        'explorer',
        'Искатель Тайн',
        'Secret Seeker',
        'Использовал поиск 50 раз',
        'Used search 50 times',
        '🔍',
        'discovery',
        '{"type": "count", "entity": "search", "action": "execute", "threshold": 50}',
        80,
        true
    ),
    (
        'sharer',
        'Даритель Света',
        'Light Giver',
        'Поделился заметкой с другим пользователем',
        'Shared a note with another user',
        '📤',
        'social',
        '{"type": "count", "entity": "share", "action": "create", "threshold": 1}',
        25,
        false
    )
ON CONFLICT (code) DO NOTHING;

COMMIT;
