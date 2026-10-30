-- Seed base achievements
-- Run with psql: psql -h <host> -U <user> -d <db> -f scripts/seed_achievements.sql

INSERT INTO achievements (code, name_ru, name_en, description_ru, description_en, icon_emoji, category, condition_json, points, hidden, created_at, updated_at)
VALUES
('first_note', 'Первая заметка', 'First note', 'Создал первую заметку', 'Created the first note', '✳️', 'engagement', '{"type":"count","entity":"note","action":"create","threshold":1}', 5, false, now(), now()),
('five_notes', 'Пять заметок', 'Five notes', 'Создал 5 заметок', 'Created 5 notes', '📝', 'engagement', '{"type":"count","entity":"note","action":"create","threshold":5}', 10, false, now(), now()),
('first_link', 'Первая связь', 'First link', 'Создал первую связь между заметками', 'Created the first link', '🔗', 'engagement', '{"type":"count","entity":"link","action":"create","threshold":1}', 5, false, now(), now()),
('share_note', 'Поделился заметкой', 'Shared a note', 'Поделился заметкой с другим пользователем', 'Shared a note with another user', '📤', 'social', '{"type":"count","entity":"share","action":"create","threshold":1}', 8, false, now(), now())
ON CONFLICT (code) DO UPDATE SET
  name_ru = EXCLUDED.name_ru,
  name_en = EXCLUDED.name_en,
  description_ru = EXCLUDED.description_ru,
  description_en = EXCLUDED.description_en,
  icon_emoji = EXCLUDED.icon_emoji,
  category = EXCLUDED.category,
  condition_json = EXCLUDED.condition_json,
  points = EXCLUDED.points,
  hidden = EXCLUDED.hidden,
  updated_at = now();
