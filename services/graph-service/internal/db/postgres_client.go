package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Note struct {
	ID    string
	Title string
	Type  string
}

type Link struct {
	Source     string
	Target     string
	LinkType   string
	Weight     float64
	SourceType string
}

type Embedding struct {
	NoteID    string
	Vector    []float32
	UpdatedAt time.Time
}

type PostgresClient interface {
	GetNotes(ctx context.Context, rootID string, depth int) ([]*Note, []*Link, error)
	GetEmbeddings(ctx context.Context, noteIDs []string) (map[string][]float32, error)
}

type postgresClient struct {
	pool *pgxpool.Pool
}

func NewPostgresClient(pool *pgxpool.Pool) PostgresClient {
	return &postgresClient{pool: pool}
}

func (c *postgresClient) GetNotes(ctx context.Context, rootID string, depth int) ([]*Note, []*Link, error) {
	if rootID == "" || depth <= 0 {
		return c.loadAll(ctx)
	}

	query := `WITH RECURSIVE nodes AS (
    SELECT id, title, type, 1 AS level
    FROM notes
    WHERE id = $1 AND deleted_at IS NULL
  UNION ALL
    SELECT n.id, n.title, n.type, nodes.level + 1
    FROM links l
    JOIN notes n ON n.id = l.target_note_id AND n.deleted_at IS NULL
    JOIN nodes ON l.source_note_id = nodes.id
    WHERE nodes.level < $2 AND l.deleted_at IS NULL
  )
  SELECT id, title, type FROM nodes;`

	rows, err := c.pool.Query(ctx, query, rootID, depth)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	notes := make([]*Note, 0)
	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Type); err != nil {
			return nil, nil, err
		}
		notes = append(notes, &note)
	}
	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}

	ids := make([]string, 0, len(notes))
	for _, note := range notes {
		ids = append(ids, note.ID)
	}
	if len(ids) == 0 {
		return notes, nil, nil
	}

	links, err := c.loadLinksByNoteIDs(ctx, ids)
	return notes, links, err
}

func (c *postgresClient) loadAll(ctx context.Context) ([]*Note, []*Link, error) {
	notesRows, err := c.pool.Query(ctx, `SELECT id, title, type FROM notes WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, nil, err
	}
	defer notesRows.Close()

	notes := make([]*Note, 0)
	for notesRows.Next() {
		var note Note
		if err := notesRows.Scan(&note.ID, &note.Title, &note.Type); err != nil {
			return nil, nil, err
		}
		notes = append(notes, &note)
	}
	if notesRows.Err() != nil {
		return nil, nil, notesRows.Err()
	}

	linksRows, err := c.pool.Query(ctx, `SELECT source_note_id, target_note_id, link_type, weight, COALESCE(source_type, 'user') FROM links WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, nil, err
	}
	defer linksRows.Close()

	links := make([]*Link, 0)
	for linksRows.Next() {
		var link Link
		if err := linksRows.Scan(&link.Source, &link.Target, &link.LinkType, &link.Weight, &link.SourceType); err != nil {
			return nil, nil, err
		}
		links = append(links, &link)
	}
	if linksRows.Err() != nil {
		return nil, nil, linksRows.Err()
	}

	return notes, links, nil
}

func (c *postgresClient) loadLinksByNoteIDs(ctx context.Context, ids []string) ([]*Link, error) {
	query := `SELECT source_note_id, target_note_id, link_type, weight, COALESCE(source_type, 'user') FROM links WHERE deleted_at IS NULL AND (source_note_id = ANY($1) OR target_note_id = ANY($1))`
	rows, err := c.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]*Link, 0)
	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.Source, &link.Target, &link.LinkType, &link.Weight, &link.SourceType); err != nil {
			return nil, err
		}
		links = append(links, &link)
	}
	if rows.Err() != nil {
		return nil, err
	}
	return links, nil
}

func (c *postgresClient) GetEmbeddings(ctx context.Context, noteIDs []string) (map[string][]float32, error) {
	if len(noteIDs) == 0 {
		return make(map[string][]float32), nil
	}

	query := `SELECT note_id, embedding, updated_at FROM note_embeddings WHERE note_id = ANY($1)`
	rows, err := c.pool.Query(ctx, query, noteIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	embeddings := make(map[string][]float32)
	for rows.Next() {
		var noteID string
		var vector []float32
		var updatedAt time.Time
		if err := rows.Scan(&noteID, &vector, &updatedAt); err != nil {
			return nil, err
		}
		embeddings[noteID] = vector
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return embeddings, nil
}
