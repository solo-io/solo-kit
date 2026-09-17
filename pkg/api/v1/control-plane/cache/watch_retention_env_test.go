package cache

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

type warningRecorder struct{ warnings []string }

func (*warningRecorder) Debugf(string, ...interface{}) {}
func (*warningRecorder) Infof(string, ...interface{})  {}
func (*warningRecorder) Errorf(string, ...interface{}) {}
func (l *warningRecorder) Warnf(format string, args ...interface{}) {
	l.warnings = append(l.warnings, fmt.Sprintf(format, args...))
}

func TestWatchRetentionEnvironment(t *testing.T) {
	for _, tc := range []struct {
		value   string
		enabled bool
		warning bool
	}{
		{"", retainUnansweredXDSWatchesDefault, false},
		{"default", retainUnansweredXDSWatchesDefault, false},
		{"on", true, false},
		{"off", false, false},
		{"invalid", retainUnansweredXDSWatchesDefault, true},
		{"ON", retainUnansweredXDSWatchesDefault, true},
		{" ", retainUnansweredXDSWatchesDefault, true},
	} {
		t.Run(fmt.Sprintf("value=%q", tc.value), func(t *testing.T) {
			t.Setenv(RetainUnansweredXDSWatchesEnv, tc.value)
			logger := &warningRecorder{}
			c := NewSnapshotCache(CacheSettings{Logger: logger}).(*snapshotCache)
			if c.retainUnansweredWatches != tc.enabled {
				t.Fatalf("retention = %t, want %t", c.retainUnansweredWatches, tc.enabled)
			}
			wantWarnings := 0
			if tc.warning {
				wantWarnings = 1
			}
			if len(logger.warnings) != wantWarnings {
				t.Fatalf("warnings = %v, want %d", logger.warnings, wantWarnings)
			}
			if tc.warning && (!strings.Contains(logger.warnings[0], RetainUnansweredXDSWatchesEnv) || !strings.Contains(logger.warnings[0], fmt.Sprintf("%q", tc.value))) {
				t.Fatalf("warning does not identify invalid setting: %s", logger.warnings[0])
			}
		})
	}
}

func TestWatchRetentionEnvironmentIsReadAtConstruction(t *testing.T) {
	t.Setenv(RetainUnansweredXDSWatchesEnv, "on")
	enabled := NewSnapshotCache(CacheSettings{}).(*snapshotCache)
	t.Setenv(RetainUnansweredXDSWatchesEnv, "off")
	disabled := NewSnapshotCache(CacheSettings{}).(*snapshotCache)
	if !enabled.retainUnansweredWatches || disabled.retainUnansweredWatches {
		t.Fatal("each cache must keep the setting read at construction")
	}
}

func TestWatchRetentionInvalidEnvironmentWithoutLogger(t *testing.T) {
	t.Setenv(RetainUnansweredXDSWatchesEnv, "invalid")
	c := NewSnapshotCache(CacheSettings{}).(*snapshotCache)
	if c.retainUnansweredWatches != retainUnansweredXDSWatchesDefault {
		t.Fatal("invalid setting must use the release default")
	}
}

func TestWatchRetentionUnsetEnvironment(t *testing.T) {
	t.Setenv(RetainUnansweredXDSWatchesEnv, "")
	if err := os.Unsetenv(RetainUnansweredXDSWatchesEnv); err != nil {
		t.Fatal(err)
	}
	logger := &warningRecorder{}
	c := NewSnapshotCache(CacheSettings{Logger: logger}).(*snapshotCache)
	if c.retainUnansweredWatches != retainUnansweredXDSWatchesDefault || len(logger.warnings) != 0 {
		t.Fatal("unset setting must use the release default without warning")
	}
}
