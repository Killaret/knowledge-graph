-- Remove the Knowledge Core system note
DELETE FROM notes WHERE id = '00000000-0000-0000-0000-000000000001'::uuid;
