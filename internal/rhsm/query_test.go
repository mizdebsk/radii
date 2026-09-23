package rhsm

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

func TestQueryRepositories(t *testing.T) {
	const base = "rhel-10-for-x86_64-baseos-rpms"
	const third = "rhel-10-for-x86_64-thirdchannel-rpms"
	tests := []struct {
		name        string
		content     string
		missingFile bool
		channels    []string
		want        []string
		missing     []string
	}{
		{name: "EnabledAndDisabled", content: "[" + base + "]\nenabled=1\n[" + third + "]\nenabled=0\n", channels: []string{"BaseOS", "ThirdChannel"}, want: []string{base, third}},
		{name: "NoEnabledFlag", content: "[" + third + "]\nname=test\n", channels: []string{"ThirdChannel"}, want: []string{third}},
		{name: "Empty", channels: []string{"BaseOS", "ThirdChannel"}, missing: []string{base, third}},
		{name: "MissingFile", missingFile: true, channels: []string{"BaseOS"}, missing: []string{base}},
		{name: "PartialDefinitions", content: "[" + base + "]\nenabled=1\n", channels: []string{"BaseOS", "ThirdChannel"}, missing: []string{third}},
		{name: "NoChannels", missingFile: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "redhat.repo")
			if !tt.missingFile {
				if err := os.WriteFile(path, []byte(tt.content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			rm := repoMgr{rhsmEnabled: true, systemInfo: sysinfo.SysInfo{IsRhel: true, OsVersion: 10, Arch: "x86_64"}, redhatRepoPath: path,
				executor: mocks.NewMockExecutor(gomock.NewController(t)), rhsmExecPath: "testdata/rhsm-exec"}
			got, err := rm.GetRepoIDs(tt.channels)
			if len(tt.missing) > 0 {
				if err == nil {
					t.Fatal("expected missing repository error")
				}
				for _, fragment := range append(tt.missing, path, "definitions are missing", "https://", "--skip-subscriptions") {
					if !strings.Contains(err.Error(), fragment) {
						t.Errorf("error %q does not contain %q", err, fragment)
					}
				}
				if strings.Contains(err.Error(), "failed to enable") {
					t.Errorf("read-only check reported an enablement failure: %v", err)
				}
			} else if err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetRepoIDs() = %v, %v; want %v", got, err, tt.want)
			}
			if !tt.missingFile {
				after, err := os.ReadFile(path)
				if err != nil || string(after) != tt.content {
					t.Fatal("query changed repository definitions")
				}
			}
		})
	}
}

func TestQueryRepositoriesBypassRHSM(t *testing.T) {
	for _, skip := range []bool{false, true} {
		name := "NonRHEL"
		if skip {
			name = "SkipSubscriptions"
		}
		t.Run(name, func(t *testing.T) {
			rm := repoMgr{rhsmEnabled: true, systemInfo: sysinfo.SysInfo{IsRhel: skip}, redhatRepoPath: t.TempDir()}
			if skip {
				rm.SetSubscriptionsEnabled(false)
			}
			got, err := rm.GetRepoIDs([]string{"BaseOS"})
			if err != nil || len(got) != 0 {
				t.Fatalf("GetRepoIDs() = %v, %v", got, err)
			}
		})
	}
}

func TestQueryRepositoriesWithoutSubscriptionManager(t *testing.T) {
	for _, tt := range []struct {
		name string
		path string
	}{
		{"missing repositories", filepath.Join(t.TempDir(), "redhat.repo")},
		{"existing repositories", "testdata/rhel10.repo"},
		{"unreadable repositories", t.TempDir()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rm := repoMgr{
				rhsmEnabled:    true,
				systemInfo:     sysinfo.SysInfo{IsRhel: true, OsVersion: 10, Arch: "x86_64"},
				redhatRepoPath: tt.path,
				rhsmExecPath:   filepath.Join(t.TempDir(), "subscription-manager"),
				executor:       mocks.NewMockExecutor(gomock.NewController(t)),
			}
			repos, err := rm.GetRepoIDs([]string{"BaseOS", "AppStream"})
			if err != nil || len(repos) != 0 {
				t.Fatalf("GetRepoIDs() = %v, %v; want no repository overrides and no error", repos, err)
			}
		})
	}
}

func TestRepositoryReadErrors(t *testing.T) {
	rm := repoMgr{rhsmEnabled: true, systemInfo: sysinfo.SysInfo{IsRhel: true, Arch: "x86_64"}, redhatRepoPath: t.TempDir(), rhsmExecPath: "testdata/rhsm-exec"}
	_, err := rm.GetRepoIDs([]string{"BaseOS"})
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("expected underlying file error, got %v", err)
	}
	if err := rm.ensureChannelsEnabled([]string{"BaseOS"}); !errors.As(err, &pathErr) {
		t.Fatalf("expected underlying file error, got %v", err)
	}
	path := filepath.Join(t.TempDir(), "redhat.repo")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", 70*1024)), 0600); err != nil {
		t.Fatal(err)
	}
	rm.redhatRepoPath = path
	if _, err := rm.GetRepoIDs([]string{"BaseOS"}); err == nil || !strings.Contains(err.Error(), "token too long") {
		t.Fatalf("expected scanner error, got %v", err)
	}
}
