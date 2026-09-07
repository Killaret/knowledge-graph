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
	assert.Contains(t, path, "backup-personal-2026-08-06-120000.sql.gz")

	// Raw sql should be removed after compression.
	raw := filepath.Join(tmp, "backup-personal-2026-08-06-120000.sql")
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
	assert.Equal(t, "backup-personal-2026-08-06-120000.sql.gz", u.uploaded[0])
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

	oldFile := filepath.Join(tmp, "backup-personal-2026-07-01-120000.sql.gz")
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
