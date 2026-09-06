package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Note struct {
	ID      string
	Title   string
	Type    string
	Creator string
	Public  bool
}

type Link struct {
	Source     string
	Target     string
	LinkType   string
	Weight     float64
	SourceType string
	Creator    string
}

type Embedding struct {
	NoteID    string
	Vector    []float32
	UpdatedAt time.Time
}

// NotesFilter defines the visibility and traversal scope for graph queries.
type NotesFilter struct {
	UserID   string // authenticated user id; empty means unscoped/internal/SKIP_AUTH
	IsPublic bool   // if true, only public notes
	RootID   string // empty means full graph
	Depth    int    // used only when RootID is set
}

type Neighbor struct {
	ID       string
	Title    string
	Type     string
	Weight   float64
	Distance int
}

type RecommendationCandidate struct {
	ID       string
	Title    string
	Type     string
	Weight   float64
	Distance int
}

type PostgresClient interface {
	GetNotes(ctx context.Context, filter NotesFilter) ([]*Note, []*Link, error)
	GetEmbeddings(ctx context.Context, noteIDs []string) (map[string][]float32, error)
	GetNoteNeighbors(ctx context.Context, filter NotesFilter, noteID string, depth int) ([]*Neighbor, error)
	GetShortestPath(ctx context.Context, filter NotesFilter, fromID, toID string) ([]string, int, float64, error)
	GetRecommendationCandidates(ctx context.Context, filter NotesFilter, noteID string, depth, limit int) ([]*RecommendationCandidate, error)
	RefreshClosureView(ctx context.Context) error
}

type postgresClient struct {
	pool      *pgxpool.Pool
	modelName string
}

// NewPostgresClient creates a PostgreSQL client for the graph service.
// modelName is the embedding model used to filter vectors, avoiding mixed
// incompatible vector spaces.
func NewPostgresClient(pool *pgxpool.Pool, modelName string) PostgresClient {
	if modelName == "" {
		modelName = "all-MiniLM-L6-v2"
	}
	return &postgresClient{pool: pool, modelName: modelName}
}

func (c *postgresClient) GetNotes(ctx context.Context, filter NotesFilter) ([]*Note, []*Link, error) {
	if filter.RootID == "" || filter.Depth <= 0 {
		return c.loadAll(ctx, filter)
	}
	return c.loadRooted(ctx, filter)
}

func (c *postgresClient) loadRooted(ctx context.Context, filter NotesFilter) ([]*Note, []*Link, error) {
	noteVis, noteArgs := noteVisibilitySQLWithAlias(filter, 2, "n")
	query := fmt.Sprintf(`WITH RECURSIVE nodes AS (
    SELECT n.id, n.title, n.type, 1 AS level
    FROM notes n
    WHERE n.id = $1 AND n.deleted_at IS NULL AND %s
  UNION ALL
    SELECT n.id, n.title, n.type, nodes.level + 1
    FROM links l
    JOIN notes n ON n.id = l.target_note_id AND n.deleted_at IS NULL AND %s
    JOIN nodes ON l.source_note_id = nodes.id
    WHERE nodes.level < $2 AND l.deleted_at IS NULL
  )
  SELECT id, title, type FROM nodes;`, noteVis, noteVis)

	args := []interface{}{filter.RootID, filter.Depth}
	args = append(args, noteArgs...)

	rows, err := c.pool.Query(ctx, query, args...)
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

	links, err := c.loadLinksByNoteIDs(ctx, ids, filter)
	return notes, links, err
}

func (c *postgresClient) loadAll(ctx context.Context, filter NotesFilter) ([]*Note, []*Link, error) {
	noteVis, noteArgs := noteVisibilitySQL(filter, 0)
	notesQuery := fmt.Sprintf(`SELECT id, title, type FROM notes WHERE deleted_at IS NULL AND %s`, noteVis)

	noteArgSlice := make([]interface{}, 0, len(noteArgs))
	noteArgSlice = append(noteArgSlice, noteArgs...)

	notesRows, err := c.pool.Query(ctx, notesQuery, noteArgSlice...)
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

	linkVis, linkArgs := linkVisibilitySQL(filter, 0)
	linksQuery := fmt.Sprintf(`SELECT source_note_id, target_note_id, link_type, weight, COALESCE(source_type, 'user') FROM links WHERE deleted_at IS NULL AND %s`, linkVis)

	linkArgSlice := make([]interface{}, 0, len(linkArgs))
	linkArgSlice = append(linkArgSlice, linkArgs...)

	linksRows, err := c.pool.Query(ctx, linksQuery, linkArgSlice...)
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

func (c *postgresClient) loadLinksByNoteIDs(ctx context.Context, ids []string, filter NotesFilter) ([]*Link, error) {
	linkVis, linkArgs := linkVisibilitySQL(filter, 1)
	query := fmt.Sprintf(`SELECT source_note_id, target_note_id, link_type, weight, COALESCE(source_type, 'user') FROM links WHERE deleted_at IS NULL AND (source_note_id = ANY($1) OR target_note_id = ANY($1)) AND %s`, linkVis)

	args := []interface{}{ids}
	args = append(args, linkArgs...)

	rows, err := c.pool.Query(ctx, query, args...)
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
		return nil, rows.Err()
	}
	return links, nil
}

func noteVisibilitySQL(filter NotesFilter, paramOffset int) (string, []interface{}) {
	if filter.IsPublic {
		return "is_public = true", nil
	}
	if filter.UserID != "" {
		return fmt.Sprintf("creator_id = $%d", paramOffset+1), []interface{}{filter.UserID}
	}
	return "TRUE", nil
}

func linkVisibilitySQL(filter NotesFilter, paramOffset int) (string, []interface{}) {
	if filter.IsPublic {
		// Only links between public notes. The sub-query is safe: ids are UUIDs,
		// visibility is a controlled boolean and the outer query already filters
		// by source/target from the rooted traversal.
		vis := "source_note_id IN (SELECT id FROM notes WHERE deleted_at IS NULL AND is_public = true) AND target_note_id IN (SELECT id FROM notes WHERE deleted_at IS NULL AND is_public = true)"
		return vis, nil
	}
	if filter.UserID != "" {
		return fmt.Sprintf("creator_id = $%d", paramOffset+1), []interface{}{filter.UserID}
	}
	return "TRUE", nil
}

func (c *postgresClient) GetEmbeddings(ctx context.Context, noteIDs []string) (map[string][]float32, error) {
	if len(noteIDs) == 0 {
		return make(map[string][]float32), nil
	}

	query := `SELECT note_id, embedding::text, updated_at FROM note_embeddings WHERE note_id = ANY($1) AND model_name = $2`
	rows, err := c.pool.Query(ctx, query, noteIDs, c.modelName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	embeddings := make(map[string][]float32)
	for rows.Next() {
		var noteID string
		var vectorText string
		var updatedAt time.Time
		if err := rows.Scan(&noteID, &vectorText, &updatedAt); err != nil {
			return nil, err
		}
		vector, err := parseVectorText(vectorText)
		if err != nil {
			return nil, err
		}
		embeddings[noteID] = vector
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return embeddings, nil
}

func parseVectorText(s string) ([]float32, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "NULL" {
		return nil, nil
	}
	s = strings.Trim(s, "[]")
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	vector := make([]float32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseFloat(p, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse vector element %q: %w", p, err)
		}
		vector = append(vector, float32(v))
	}
	return vector, nil
}

func (c *postgresClient) GetNoteNeighbors(ctx context.Context, filter NotesFilter, noteID string, depth int) ([]*Neighbor, error) {
	if depth <= 0 {
		depth = 2
	}

	vis, visArgs := noteVisibilitySQLWithAlias(filter, 2, "n")
	query := fmt.Sprintf(`SELECT id, title, type, distance, weight FROM (
	  SELECT c.descendant_id AS id, n.title, n.type, c.distance, c.weight
	  FROM note_links_closure c
	  JOIN notes n ON n.id = c.descendant_id
	  WHERE c.ancestor_id = $1 AND c.distance <= $2 AND n.deleted_at IS NULL AND %s
	  UNION
	  SELECT c.ancestor_id AS id, n.title, n.type, c.distance, c.weight
	  FROM note_links_closure c
	  JOIN notes n ON n.id = c.ancestor_id
	  WHERE c.descendant_id = $1 AND c.distance <= $2 AND n.deleted_at IS NULL AND %s
	) AS neighbors
	ORDER BY distance ASC, weight DESC`, vis, vis)

	args := []interface{}{noteID, depth}
	args = append(args, visArgs...)

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	neighbors := make([]*Neighbor, 0)
	for rows.Next() {
		var n Neighbor
		if err := rows.Scan(&n.ID, &n.Title, &n.Type, &n.Distance, &n.Weight); err != nil {
			return nil, err
		}
		if n.ID == noteID {
			continue
		}
		neighbors = append(neighbors, &n)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return neighbors, nil
}

func (c *postgresClient) GetShortestPath(ctx context.Context, filter NotesFilter, fromID, toID string) ([]string, int, float64, error) {
	if fromID == "" || toID == "" {
		return nil, 0, 0, fmt.Errorf("from_id and to_id are required")
	}

	vis, visArgs := noteVisibilitySQLForEndpoints(filter, 2)
	query := fmt.Sprintf(`SELECT array_to_string(c.path, ','), c.distance, c.weight
	FROM note_links_closure c
	JOIN notes nf ON nf.id = c.ancestor_id
	JOIN notes nt ON nt.id = c.descendant_id
	WHERE c.ancestor_id = $1 AND c.descendant_id = $2
	  AND nf.deleted_at IS NULL AND nt.deleted_at IS NULL AND %s
	LIMIT 1`, vis)

	args := []interface{}{fromID, toID}
	args = append(args, visArgs...)

	var pathStr string
	var distance int
	var weight float64
	err := c.pool.QueryRow(ctx, query, args...).Scan(&pathStr, &distance, &weight)
	if err == nil && pathStr != "" {
		return strings.Split(pathStr, ","), distance, weight, nil
	}

	// Try reverse direction and reverse the returned path.
	vis2, visArgs2 := noteVisibilitySQLForEndpoints(filter, 2)
	query2 := fmt.Sprintf(`SELECT array_to_string(c.path, ','), c.distance, c.weight
	FROM note_links_closure c
	JOIN notes nf ON nf.id = c.ancestor_id
	JOIN notes nt ON nt.id = c.descendant_id
	WHERE c.ancestor_id = $2 AND c.descendant_id = $1
	  AND nf.deleted_at IS NULL AND nt.deleted_at IS NULL AND %s
	LIMIT 1`, vis2)

	args2 := []interface{}{fromID, toID}
	args2 = append(args2, visArgs2...)

	err = c.pool.QueryRow(ctx, query2, args2...).Scan(&pathStr, &distance, &weight)
	if err != nil {
		return nil, 0, 0, err
	}
	parts := strings.Split(pathStr, ",")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return parts, distance, weight, nil
}

func (c *postgresClient) GetRecommendationCandidates(ctx context.Context, filter NotesFilter, noteID string, depth, limit int) ([]*RecommendationCandidate, error) {
	if depth <= 0 {
		depth = 2
	}
	if limit <= 0 {
		limit = 10
	}

	vis, visArgs := noteVisibilitySQLWithAlias(filter, 3, "n")
	query := fmt.Sprintf(`SELECT n.id, n.title, n.type, c.distance, c.weight
	FROM note_links_closure c
	JOIN notes n ON n.id = c.descendant_id
	WHERE c.ancestor_id = $1 AND c.distance <= $2 AND c.descendant_id <> $1
	  AND n.deleted_at IS NULL AND %s
	ORDER BY c.distance ASC, c.weight DESC
	LIMIT $3`, vis)

	args := []interface{}{noteID, depth, limit}
	args = append(args, visArgs...)

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]*RecommendationCandidate, 0)
	for rows.Next() {
		var r RecommendationCandidate
		if err := rows.Scan(&r.ID, &r.Title, &r.Type, &r.Distance, &r.Weight); err != nil {
			return nil, err
		}
		candidates = append(candidates, &r)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return candidates, nil
}

func (c *postgresClient) RefreshClosureView(ctx context.Context) error {
	_, err := c.pool.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY note_links_closure")
	return err
}

func noteVisibilitySQLWithAlias(filter NotesFilter, paramOffset int, alias string) (string, []interface{}) {
	if filter.IsPublic {
		return fmt.Sprintf("%s.is_public = true", alias), nil
	}
	if filter.UserID != "" {
		return fmt.Sprintf("%s.creator_id = $%d", alias, paramOffset+1), []interface{}{filter.UserID}
	}
	return "TRUE", nil
}

func noteVisibilitySQLForEndpoints(filter NotesFilter, paramOffset int) (string, []interface{}) {
	if filter.IsPublic {
		return "nf.is_public = true AND nt.is_public = true", nil
	}
	if filter.UserID != "" {
		return fmt.Sprintf("nf.creator_id = $%d AND nt.creator_id = $%d", paramOffset+1, paramOffset+1), []interface{}{filter.UserID}
	}
	return "TRUE", nil
}
