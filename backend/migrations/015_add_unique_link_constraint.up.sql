-- Добавляем уникальное ограничение на связи между заметками
-- Одна пара заметок может иметь только одну связь одного типа
ALTER TABLE links ADD CONSTRAINT links_source_target_type_unique UNIQUE (source_note_id, target_note_id, link_type);
