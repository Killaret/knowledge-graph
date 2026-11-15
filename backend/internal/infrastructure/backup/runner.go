package backup

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Uploader uploads a local backup file to remote storage.
type Uploader interface {
	UploadBackup(ctx context.Context, localPath, remoteKey string) error
}

// Runner creates a PostgreSQL dump, compresses it and optionally uploads it.
type Runner struct {
	dbURL         string
	localPath     string
	remoteFolder  string
	retention     int
	uploader      Uploader
	pgDumpExec    func(ctx context.Context, dsn, outFile string) error
	mongoURI      string
	mongoDB       string
	mongoDumpExec func(ctx context.Context, uri, db, outFile string) error
	clock         func() time.Time
}

// Option configures a Runner.
type Option func(*Runner)

// WithUploader sets the uploader used for cloud backups.
func WithUploader(u Uploader) Option {
	return func(r *Runner) { r.uploader = u }
}

// WithRemoteFolder sets the folder (prefix) used for the remote backup key.
func WithRemoteFolder(folder string) Option {
	return func(r *Runner) { r.remoteFolder = folder }
}

// WithPgDumpExec replaces the default pg_dump execution for testing.
func WithPgDumpExec(fn func(context.Context, string, string) error) Option {
	return func(r *Runner) { r.pgDumpExec = fn }
}

// WithMongo enables a mongodump archive next to every SQL dump (BACKUP-4):
// nlp_artifacts and quality_log live in MongoDB and were not backed up.
func WithMongo(uri, db string) Option {
	return func(r *Runner) { r.mongoURI, r.mongoDB = uri, db }
}

// WithMongoDumpExec replaces the default mongodump execution for testing.
func WithMongoDumpExec(fn func(context.Context, string, string, string) error) Option {
	return func(r *Runner) { r.mongoDumpExec = fn }
}

// WithClock replaces the clock for testing.
func WithClock(fn func() time.Time) Option {
	return func(r *Runner) { r.clock = fn }
}

// NewRunner creates a new database backup runner.
func NewRunner(dbURL, localPath string, retention int, opts ...Option) *Runner {
	r := &Runner{
		dbURL:     dbURL,
		localPath: localPath,
		retention: retention,
		clock:     time.Now,
		pgDumpExec: func(ctx context.Context, dsn, out string) error {
			cmd := exec.CommandContext(ctx, "pg_dump", dsn, "-f", out, "--format=plain", "--no-owner", "--no-acl")
			outb, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("pg_dump failed: %w: %s", err, string(outb))
			}
			return nil
		},
		mongoDumpExec: func(ctx context.Context, uri, db, out string) error {
			cmd := exec.CommandContext(ctx, "mongodump", "--uri="+uri, "--db="+db, "--archive="+out, "--gzip")
			outb, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("mongodump failed: %w: %s", err, string(outb))
			}
			return nil
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Run performs the database backup and returns the path to the compressed file.
func (r *Runner) Run(ctx context.Context) (string, error) {
	if err := os.MkdirAll(r.localPath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create backup dir: %w", err)
	}

	ts := r.clock().UTC().Format("2006-01-02-150405")
	// Event-driven dumps get their own name space: the shared backup folder
	// also holds daily/weekly snapshots from the cron scheduler, and the
	// worker's short retention must not eat files with a longer lifetime.
	baseName := fmt.Sprintf("backup-personal-auto-%s", ts)
	sqlFile := filepath.Join(r.localPath, baseName+".sql")
	gzFile := sqlFile + ".gz"

	if err := r.pgDumpExec(ctx, r.dbURL, sqlFile); err != nil {
		return "", err
	}
	defer os.Remove(sqlFile)

	if err := r.gzipFile(sqlFile, gzFile); err != nil {
		return "", err
	}

	// MongoDB archive beside the SQL dump (BACKUP-4 item 2). A failed mongo
	// dump fails the run: a backup that silently skips a datastore is the
	// bug this change exists to remove.
	if r.mongoURI != "" {
		mongoFile := filepath.Join(r.localPath, baseName+".mongo.gz")
		if err := r.mongoDumpExec(ctx, r.mongoURI, r.mongoDB, mongoFile); err != nil {
			return "", err
		}
	}

	if r.uploader != nil {
		uploads := []string{gzFile}
		if r.mongoURI != "" {
			uploads = append(uploads, filepath.Join(r.localPath, baseName+".mongo.gz"))
		}
		for _, local := range uploads {
			remoteKey := filepath.Base(local)
			if r.remoteFolder != "" {
				remoteKey = path.Join(r.remoteFolder, remoteKey)
			}
			if err := r.uploader.UploadBackup(ctx, local, remoteKey); err != nil {
				return "", fmt.Errorf("upload failed: %w", err)
			}
		}
	}

	if err := r.cleanupOldBackups(); err != nil {
		log.Printf("[Backup] cleanup warning: %v", err)
	}

	return gzFile, nil
}

func (r *Runner) gzipFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open raw backup: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create compressed backup: %w", err)
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	defer gz.Close()

	if _, err := io.Copy(gz, in); err != nil {
		return fmt.Errorf("gzip copy failed: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("gzip close failed: %w", err)
	}
	return nil
}

func (r *Runner) cleanupOldBackups() error {
	if r.retention <= 0 {
		return nil
	}
	entries, err := os.ReadDir(r.localPath)
	if err != nil {
		return err
	}
	cutoff := r.clock().Add(-time.Duration(r.retention) * 24 * time.Hour)
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "backup-personal-auto-") ||
			(!strings.HasSuffix(e.Name(), ".sql.gz") && !strings.HasSuffix(e.Name(), ".mongo.gz")) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(r.localPath, e.Name()))
		}
	}
	return nil
}
