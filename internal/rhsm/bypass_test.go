package rhsm

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestRepositoriesBypassSetup(t *testing.T) {
	for _, tt := range []struct {
		name     string
		enabled  bool
		isRhel   bool
		channels []string
	}{
		{"subscriptions disabled", false, true, []string{"BaseOS"}},
		{"non-RHEL", true, false, []string{"BaseOS"}},
		{"no channels", true, true, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rm := repoMgr{
				rhsmEnabled:    tt.enabled,
				systemInfo:     sysinfo.SysInfo{IsRhel: tt.isRhel, OsVersion: 10, Arch: "x86_64"},
				executor:       mocks.NewMockExecutor(gomock.NewController(t)),
				redhatRepoPath: t.TempDir(),
				rhsmExecPath:   "testdata/rhsm-exec",
			}
			repos, err := rm.GetRepoIDs(tt.channels)
			if err != nil || len(repos) != 0 {
				t.Errorf("GetRepoIDs() = %v, %v; want no repositories and no error", repos, err)
			}
			if err := rm.EnsureRepositoriesEnabled(tt.channels); err != nil {
				t.Errorf("EnsureRepositoriesEnabled() = %v; want no error", err)
			}
		})
	}
}
