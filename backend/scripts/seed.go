package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

// Sample data for seeding
var (
	starTitles = []string{
		"Alpha Centauri", "Sirius", "Betelgeuse", "Vega", "Rigel",
		"Arcturus", "Antares", "Pollux", "Altair", "Deneb",
		"Полярная звезда", "Капелла", "Алдебаран", "Регул", "Шедар",
	}

	planetTitles = []string{
		"Earth", "Mars", "Jupiter", "Saturn", "Venus",
		"Mercury", "Uranus", "Neptune", "Pluto", "Kepler-442b",
		"Земля", "Марс", "Юпитер", "Сатурн", "Венера",
		"Меркурий", "Уран", "Нептун", "Плутон", "TRAPPIST-1e",
	}

	cometTitles = []string{
		"Halley's Comet", "Hale-Bopp", "NEOWISE", "Encke", "Hyakutake",
		"West", "Ikeya-Seki", "Comet Lovejoy", "Comet McNaught", "ISON",
		"Комета Галлея", "Комета Энке", "Комета Хякутакэ", "Комета Веста", "Комета Исона",
	}

	galaxyTitles = []string{
		"Milky Way", "Andromeda", "Triangulum", "Whirlpool", "Sombrero",
		"Pinwheel", "Cartwheel", "Black Eye", "Cigar", "Bode's Galaxy",
		"Млечный Путь", "Галактика Андромеды", "Треугольник", "Водоворот", "Сомбреро",
	}

	contentSamples = []string{
		"This astronomical object has been studied for centuries by scientists worldwide.",
		"Recent observations suggest unusual activity in the stellar atmosphere.",
		"The discovery was made using advanced telescopic equipment in 2024.",
		"Scientists believe this could be key to understanding cosmic evolution.",
		"Research indicates potential for future exploration missions.",
		"Этот астрономический объект изучается учеными на протяжении веков.",
		"Недаблюдения указывают на необычную активность в звездной атмосфере.",
		"Открытие было сделано с использованием продвинутого телескопического оборудования.",
		"Ученые считают, что это может быть ключом к пониманию космической эволюции.",
		"Исследования указывают на потенциал для будущих исследовательских миссий.",
	}
)

type Note struct {
	ID        string
	Title     string
	Content   string
	Type      string
	CreatedAt time.Time
}

type Link struct {
	SourceNoteID string
	TargetNoteID string
	LinkType     string
	Weight       float64
}

func main() {
	ctx := context.Background()

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/knowledge_graph"
	}

	// Connect to database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	fmt.Println("🌌 Knowledge Graph Seeder")
	fmt.Println("========================")

	// Clear existing data
	fmt.Println("\n🧹 Clearing existing data...")
	if err := clearData(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "Error clearing data: %v\n", err)
		os.Exit(1)
	}

	// Generate notes
	fmt.Println("\n📝 Generating notes...")
	notes := generateNotes(75)

	// Insert notes
	noteIDs, err := insertNotes(ctx, pool, notes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inserting notes: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Created %d notes\n", len(noteIDs))

	// Generate links
	fmt.Println("\n🔗 Generating connections...")
	links := generateLinks(noteIDs, 100)

	// Insert links
	linkCount, err := insertLinks(ctx, pool, links)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inserting links: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Created %d connections\n", linkCount)

	fmt.Println("\n✨ Seeding completed successfully!")
	fmt.Printf("📊 Generated: %d notes, %d connections\n", len(noteIDs), linkCount)
}

func clearData(ctx context.Context, pool *pgxpool.Pool) error {
	queries := []string{
		"DELETE FROM links",
		"DELETE FROM note_keywords",
		"DELETE FROM notes",
	}

	for _, query := range queries {
		if _, err := pool.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func generateNotes(count int) []Note {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	notes := make([]Note, 0, count)

	types := []struct {
		name   string
		titles []string
	}{
		{"star", starTitles},
		{"planet", planetTitles},
		{"satellite", planetTitles},
		{"comet", cometTitles},
		{"galaxy", galaxyTitles},
		{"asteroid", cometTitles}, // Астероиды
		{"debris", starTitles},    // Обломки
	}

	for i := 0; i < count; i++ {
		typeInfo := types[r.Intn(len(types))]
		title := typeInfo.titles[r.Intn(len(typeInfo.titles))]

		// Make title unique by adding number if needed
		if i >= len(typeInfo.titles) {
			title = fmt.Sprintf("%s %d", title, i)
		}

		content := contentSamples[r.Intn(len(contentSamples))]
		// Add more content
		for j := 0; j < 2; j++ {
			content += " " + contentSamples[r.Intn(len(contentSamples))]
		}

		note := Note{
			ID:        fmt.Sprintf("note_%d", i),
			Title:     title,
			Content:   content,
			Type:      typeInfo.name,
			CreatedAt: time.Now().Add(-time.Duration(r.Intn(365)) * 24 * time.Hour),
		}
		notes = append(notes, note)
	}

	return notes
}

func insertNotes(ctx context.Context, pool *pgxpool.Pool, notes []Note) ([]string, error) {
	var noteIDs []string

	for _, note := range notes {
		query := `
			INSERT INTO notes (title, content, type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $4)
			RETURNING id
		`

		var id string
		err := pool.QueryRow(ctx, query, note.Title, note.Content, note.Type, note.CreatedAt).Scan(&id)
		if err != nil {
			return nil, err
		}
		noteIDs = append(noteIDs, id)
	}

	return noteIDs, nil
}

func generateLinks(noteIDs []string, count int) []Link {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	links := make([]Link, 0, count)
	linkTypes := []string{"reference", "related", "inspired", "similar", "derived"}

	for i := 0; i < count; i++ {
		sourceIdx := r.Intn(len(noteIDs))
		targetIdx := r.Intn(len(noteIDs))

		// Ensure source and target are different
		for targetIdx == sourceIdx {
			targetIdx = r.Intn(len(noteIDs))
		}

		link := Link{
			SourceNoteID: noteIDs[sourceIdx],
			TargetNoteID: noteIDs[targetIdx],
			LinkType:     linkTypes[r.Intn(len(linkTypes))],
			Weight:       0.5 + r.Float64()*0.5, // Random weight between 0.5 and 1.0
		}
		links = append(links, link)
	}

	return links
}

func insertLinks(ctx context.Context, pool *pgxpool.Pool, links []Link) (int, error) {
	count := 0

	for _, link := range links {
		query := `
			INSERT INTO links (source_note_id, target_note_id, link_type, weight)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (source_note_id, target_note_id, link_type) DO NOTHING
		`

		result, err := pool.Exec(ctx, query, link.SourceNoteID, link.TargetNoteID, link.LinkType, link.Weight)
		if err != nil {
			// Skip duplicate link errors
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			return count, err
		}
		count += int(result.RowsAffected())
	}

	return count, nil
}
