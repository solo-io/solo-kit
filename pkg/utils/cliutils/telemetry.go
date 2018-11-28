package cliutils


import (
	"context"
	"path/filepath"
	"time"

	"github.com/solo-io/solo-kit/pkg/utils/fsutils"
	checkpoint "github.com/solo-io/go-checkpoint"
)

// Telemetry sends telemetry information about glooctl to Checkpoint server
func Telemetry(version string, t time.Time) {
	sigfile := filepath.Join(fsutils.HomeDir(), ".glooctl.sig")
	configDir, err := fsutils.ConfigDir()
	if err == nil {
		sigfile = filepath.Join(configDir, "telemetry.sig")
	}
	ctx := context.Background()
	report := &checkpoint.ReportParams{
		Product:       "glooctl",
		Version:       version,
		StartTime:     t,
		EndTime:       time.Now(),
		SignatureFile: sigfile,
	}
	checkpoint.Report(ctx, report)
}
