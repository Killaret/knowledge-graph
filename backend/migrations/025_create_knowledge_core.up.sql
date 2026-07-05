-- Create the Knowledge Core system note used for in-canvas help and documentation
INSERT INTO notes (id, title, content, type, metadata, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,
    'Knowledge Core',
    'The Knowledge Core is your in-app guide. Use it to remember keyboard shortcuts and graph gestures:

- F: search nodes by name
- Esc: toggle focus mode (hide decorative effects)
- ?: open this help modal
- Ctrl+Shift+N: quick capture a new note
- Drag node onto another node: create a link
- Drag node into the black hole: delete the note
- Double-click empty space: create a new note
- Mouse wheel / pinch: zoom in and out

Nodes are visualized as astronomical objects based on their type. New notes created within the last 24 hours are highlighted with a pulsing turquoise ring.',
    'technical',
    '{"system": true, "protected": true}',
    now(),
    now()
)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    content = EXCLUDED.content,
    type = EXCLUDED.type,
    metadata = EXCLUDED.metadata,
    updated_at = now();
