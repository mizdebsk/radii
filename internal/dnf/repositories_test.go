package dnf

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestAvailableCacheFollowsRepositoryConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	executor := mocks.NewMockExecutor(ctrl)
	pm := NewPackageManager(executor)
	const format = "QQQ|%{name}|%{epoch}|%{version}|%{release}|%{arch}|%{sourcerpm}|%{repoid}|YYY\n"
	query := func(repos []string, name string) {
		t.Helper()
		args := []string{"-q", "repoquery"}
		for _, repo := range repos {
			args = append(args, "--enablerepo="+repo)
		}
		args = append(args, "--qf", format)
		executor.EXPECT().RunCapture("dnf", args).Return([]string{"QQQ|" + name + "|0|1|1|noarch|" + name + "-1-1.src.rpm|test|YYY"}, nil).Times(1)
		for range 2 {
			got, err := pm.ListAvailablePackages()
			if err != nil || len(got) != 1 || got[0].Name != name {
				t.Fatalf("query = %v, %v; want %s", got, err, name)
			}
		}
	}
	query(nil, "initial")
	repos := []string{"repo-a", "repo-b"}
	pm.SetEnableRepos(repos)
	repos[0] = "caller-mutation"
	query([]string{"repo-a", "repo-b"}, "configured")
	pm.SetEnableRepos([]string{"repo-a", "repo-b"})
	query([]string{"repo-a", "repo-b"}, "refreshed")
	pm.SetEnableRepos(nil)
	query(nil, "cleared")
	pm = NewPackageManager(executor)
	query(nil, "separate-manager")
}
