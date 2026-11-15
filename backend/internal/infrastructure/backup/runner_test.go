package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUploader struct {
	uploaded []string
	err      error
}

func (f *fakeUploader) UploadBackup(ctx context.Context, localPath, remoteKey string) error {
	if f.err != nil {
		return f.err
	}
	f.uploaded = append(f.uploaded, remoteKey)
	return nil
}

func TestRunnerCreatesAndCompressesBackup(t *testing.T) {
	tmp := t.TempDir()

	r := NewRunner(
		"postgresql://user:pass@localhost/db",
		tmp,
		0,
		WithPgDumpExec(func(ctx context.Context, dsn, out string) error {
			assert.Equal(t, "postgresql://user:pass@localhost/db", dsn)
			return os.WriteFile(out, []byte("-- test dump"), 0o644)
		}),
		WithUploader(&fakeUploader{}),
		WithClock(func() time.Time { return time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC) }),
	)

	path, err := r.Run(context.Background())
	require.NoError(t, err)
	assert.FileExists(t, path)
	assert.Contains(t, path, "backup-personal-auto-2026-08-06-120000.sql.gz")

	// Raw sql should be removed after compression.
	raw := filepath.Join(tmp, "backup-personal-auto-2026-08-06-120000.sql")
	assert.NoFileExists(t, raw)
}

func TestRunnerUploadsBackup(t *testing.T) {
	tmp := t.TempDir()
	u := &fakeUploader{}

	r := NewRunner(
		"postgresql://user:pass@localhost/db",
		tmp,
		0,
		WithPgDumpExec(func(ctx context.Context, dsn, out string) error {
			return os.WriteFile(out, []byte("-- test dump"), 0o644)
		}),
		WithUploader(u),
		WithClock(func() time.Time { return time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC) }),
	)

	_, err := r.Run(context.Background())
	require.NoError(t, err)
	require.Len(t, u.uploaded, 1)
	assert.Equal(t, "backup-personal-auto-2026-08-06-120000.sql.gz", u.uploaded[0])
}

func TestRunnerUploadError(t *testing.T) {
	tmp := t.TempDir()
	u := &fakeUploader{err: fmt.Errorf("yandex failed")}

	r := NewRunner(
		"postgresql://user:pass@localhost/db",
		tmp,
		0,
		WithPgDumpExec(func(ctx context.Context, dsn, out string) error {
			return os.WriteFile(out, []byte("-- test dump"), 0o644)
		}),
		WithUploader(u),
		WithClock(func() time.Time { return time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC) }),
	)

	_, err := r.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "yandex failed")
}

func TestRunnerCleanupOldBackups(t *testing.T) {
	tmp := t.TempDir()
	clock := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)

	oldFile := filepath.Join(tmp, "backup-personal-auto-2026-07-01-120000.sql.gz")
	require.NoError(t, os.WriteFile(oldFile, []byte("old"), 0o644))
	oldTime := clock.Add(-30 * 24 * time.Hour)
	require.NoError(t, os.Chtimes(oldFile, oldTime, oldTime))

	r := NewRunner(
		"postgresql://user:pass@localhost/db",
		tmp,
		7,
		WithPgDumpExec(func(ctx context.Context, dsn, out string) error {
			return os.WriteFile(out, []byte("-- test dump"), 0o644)
		}),
		WithUploader(&fakeUploader{}),
		WithClock(func() time.Time { return clock }),
	)

	_, err := r.Run(context.Background())
	require.NoError(t, err)
	assert.NoFileExists(t, oldFile)
}

// The worker shares the backup folder with the cron scheduler: daily files
// live 7 days by their own rules and weekly files 90 days. The worker's short
// retention must remove only its own event-driven dumps.
func TestRunnerCleanupKeepsSchedulerBackups(t *testing.T) {
	tmp := t.TempDir()
	clock := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	oldTime := clock.Add(-30 * 24 * time.Hour)

	oldAuto := filepath.Join(tmp, "backup-personal-auto-2026-07-01-120000.sql.gz")
	oldDaily := filepath.Join(tmp, "backup-personal-daily-2026-07-01-120000.sql.gz")
	oldWeekly := filepath.Join(tmp, "backup-personal-weekly-2026-07-01-120000.sql.gz")
	for _, f := range []string{oldAuto, oldDaily, oldWeekly} {
		require.NoError(t, os.WriteFile(f, []byte("old"), 0o644))
		require.NoError(t, os.Chtimes(f, oldTime, oldTime))
	}

	r := NewRunner(
		"postgresql://user:pass@localhost/db",
		tmp,
		7,
		WithPgDumpExec(func(ctx context.Context, dsn, out string) error {
			return os.WriteFile(out, []byte("-- test dump"), 0o644)
		}),
		WithClock(func() time.Time { return clock }),
	)

	_, err := r.Run(context.Background())
	require.NoError(t, err)
	assert.NoFileExists(t, oldAuto)
	assert.FileExists(t, oldDaily)
	assert.FileExists(t, oldWeekly)
}

// BACKUP-4 item 2: every backup run must also leave a MongoDB archive.
func TestRunnerDumpsMongo(t *testing.T) {
	tmp := t.TempDir()
	var gotURI, gotDB string

	r := NewRunner(
		"postgresql://user:pass@localhost/db",
		tmp,
		0,
		WithPgDumpExec(func(ctx context.Context, dsn, out string) error {
			return os.WriteFile(out, []byte("-- test dump"), 0o644)
		}),
		WithMongo("mongodb://mongo_personal:27017", "knowledge_personal"),
		WithMongoDumpExec(func(ctx context.Context, uri, db, out string) error {
			gotURI, gotDB = uri, db
			return os.WriteFile(out, []byte("mongo-archive"), 0o644)
		}),
		WithClock(func() time.Time { return time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC) }),
	)

	_, err := r.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "mongodb://mongo_personal:27017", gotURI)
	assert.Equal(t, "knowledge_personal", gotDB)
	assert.FileExists(t, filepath.Join(tmp, "backup-personal-auto-2026-08-06-120000.mongo.gz"))
}

// A failed mongodump fails the whole run: a backup that silently skips a
// datastore is the bug this change exists to remove.
func TestRunnerMongoDumpError(t *testing.T) {
	tmp := t.TempDir()

	r := NewRunner(
		"postgresql://user:pass@localhost/db",
		tmp,
		0,
		WithPgDumpExec(func(ctx context.Context, dsn, out string) error {
			return os.WriteFile(out, []byte("-- test dump"), 0o644)
		}),
		WithMongo("mongodb://mongo_personal:27017", "knowledge_personal"),
		WithMongoDumpExec(func(ctx context.Context, uri, db, out string) error {
			return fmt.Errorf("mongodump not found")
		}),
	)

	_, err := r.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mongodump not found")
}

func TestRunnerWithoutMongoKeepsOldBehavior(t *testing.T) {
	tmp := t.TempDir()

	r := NewRunner(
		"postgresql://user:pass@localhost/db",
		tmp,
		0,
		WithPgDumpExec(func(ctx context.Context, dsn, out string) error {
			return os.WriteFile(out, []byte("-- test dump"), 0o644)
		}),
		WithClock(func() time.Time { return time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC) }),
	)

	path, err := r.Run(context.Background())
	require.NoError(t, err)
	assert.FileExists(t, path)
	matches, err := filepath.Glob(filepath.Join(tmp, "*.mongo.gz"))
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestRunnerUploadsMongoToo(t *testing.T) {
	tmp := t.TempDir()
	u := &fakeUploader{}

	r := NewRunner(
		"postgresql://user:pass@localhost/db",
		tmp,
		0,
		WithPgDumpExec(func(ctx context.Context, dsn, out string) error {
			return os.WriteFile(out, []byte("-- test dump"), 0o644)
		}),
		WithUploader(u),
		WithMongo("mongodb://mongo_personal:27017", "knowledge_personal"),
		WithMongoDumpExec(func(ctx context.Context, uri, db, out string) error {
			return os.WriteFile(out, []byte("mongo-archive"), 0o644)
		}),
		WithClock(func() time.Time { return time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC) }),
	)

	_, err := r.Run(context.Background())
	require.NoError(t, err)
	require.Len(t, u.uploaded, 2)
	assert.Equal(t, "backup-personal-auto-2026-08-06-120000.sql.gz", u.uploaded[0])
	assert.Equal(t, "backup-personal-auto-2026-08-06-120000.mongo.gz", u.uploaded[1])
}

func TestRunnerCleanupOldMongoBackups(t *testing.T) {
	tmp := t.TempDir()
	clock := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)

	oldFile := filepath.Join(tmp, "backup-personal-auto-2026-07-01-120000.mongo.gz")
	require.NoError(t, os.WriteFile(oldFile, []byte("old"), 0o644))
	oldTime := clock.Add(-30 * 24 * time.Hour)
	require.NoError(t, os.Chtimes(oldFile, oldTime, oldTime))

	r := NewRunner(
		"postgresql://user:pass@localhost/db",
		tmp,
		7,
		WithPgDumpExec(func(ctx context.Context, dsn, out string) error {
			return os.WriteFile(out, []byte("-- test dump"), 0o644)
		}),
		WithClock(func() time.Time { return clock }),
	)

	_, err := r.Run(context.Background())
	require.NoError(t, err)
	assert.NoFileExists(t, oldFile)
}
