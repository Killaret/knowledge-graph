-- Добавляем поле type для заметок (star, planet, comet, galaxy, asteroid)
ALTER TABLE notes ADD COLUMN type VARCHAR(50) NOT NULL DEFAULT 'star';

-- Индекс для быстрой фильтрации по типу
CREATE INDEX idx_notes_type ON notes(type);
