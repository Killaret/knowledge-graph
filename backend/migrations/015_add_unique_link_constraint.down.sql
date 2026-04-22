-- Удаляем уникальное ограничение на связи
ALTER TABLE links DROP CONSTRAINT IF EXISTS links_source_target_type_unique;
