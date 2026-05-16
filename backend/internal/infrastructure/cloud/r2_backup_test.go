package cloud

import (
	"context"
	"testing"
)

func TestNewR2BackupService(t *testing.T) {
	tests := []struct {
		name    string
		config  R2Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: R2Config{
				AccountID:       "test-account",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				Bucket:          "test-bucket",
				Region:          "auto",
			},
			wantErr: false,
		},
		{
			name: "missing account ID",
			config: R2Config{
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				Bucket:          "test-bucket",
				Region:          "auto",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := NewR2BackupService(ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewR2BackupService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
