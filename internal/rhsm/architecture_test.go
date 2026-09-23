package rhsm

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestRepositoriesWithUnknownArchitecture(t *testing.T) {
	for _, tt := range []struct {
		name     string
		enabled  bool
		isRhel   bool
		channels []string
		wantErr  bool
	}{
		{"required architecture", true, true, []string{"BaseOS"}, true},
		{"subscriptions disabled", false, true, []string{"BaseOS"}, false},
		{"non-RHEL", true, false, []string{"BaseOS"}, false},
		{"no channels", true, true, nil, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rm := repoMgr{
				rhsmEnabled:    tt.enabled,
				systemInfo:     sysinfo.SysInfo{IsRhel: tt.isRhel, OsVersion: 10},
				executor:       mocks.NewMockExecutor(gomock.NewController(t)),
				redhatRepoPath: t.TempDir(),
				rhsmExecPath:   "testdata/rhsm-exec",
			}
			repos, queryErr := rm.GetRepoIDs(tt.channels)
			if len(repos) != 0 {
				t.Errorf("GetRepoIDs returned repositories without an architecture: %v", repos)
			}
			enableErr := rm.EnsureRepositoriesEnabled(tt.channels)
			for operation, err := range map[string]error{"query": queryErr, "enable": enableErr} {
				if tt.wantErr {
					if err == nil || !strings.Contains(err.Error(), "system architecture is unknown") {
						t.Errorf("%s: expected unknown architecture error, got %v", operation, err)
					}
				} else if err != nil {
					t.Errorf("%s: unexpected error: %v", operation, err)
				}
			}
		})
	}
}
