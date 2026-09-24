package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestRemoveDriverSelection(t *testing.T) {
	for _, command := range []string{"remove", "rm"} {
		for _, tt := range []struct {
			argument string
			provider string
			versions []string
			selected []string
		}{
			{"nvidia", "nvidia", []string{"580.100", "580.9"}, []string{""}},
			{"nvidia", "nvidia", nil, []string{""}},
			{"nvidia:580.9", "nvidia", []string{"580.100", "580.9"}, []string{"580.9"}},
			{"amdgpu", "amdgpu", []string{"latest"}, []string{""}},
		} {
			t.Run(command+"/"+tt.argument, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				provider := mocks.NewMockProvider(ctrl)
				other := mocks.NewMockProvider(ctrl)
				pm := mocks.NewMockPackageManager(ctrl)
				provider.EXPECT().GetID().Return(tt.provider).AnyTimes()
				other.EXPECT().GetID().Return("other").AnyTimes()
				var installed, selected []api.DriverID
				for _, version := range tt.versions {
					installed = append(installed, api.DriverID{ProviderID: tt.provider, Version: version})
				}
				for _, version := range tt.selected {
					selected = append(selected, api.DriverID{ProviderID: tt.provider, Version: version})
				}
				if strings.Contains(tt.argument, ":") {
					provider.EXPECT().ListInstalled().Return(installed, nil)
				}
				provider.EXPECT().Remove(selected).Return([]string{"selected-package"}, nil)
				pm.EXPECT().Remove([]string{"selected-package"}, true, true).Return(nil)
				deps := api.CoreDeps{
					Providers: []api.Provider{other, provider}, PackageManager: pm,
					RepositoryManager: mocks.NewMockRepositoryManager(ctrl),
				}
				if err := Execute([]string{"radii", command, "--batch", "--dry-run", tt.argument}, deps, "test"); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestRemoveBareDriverErrors(t *testing.T) {
	for _, tt := range []struct {
		name     string
		argument string
		query    bool
		queryErr error
		want     string
	}{
		{"none installed", "nvidia", true, nil, "nothing to remove"},
		{"query failure", "nvidia", true, fmt.Errorf("query failed"), "failed to remove NVIDIA driver: query failed"},
		{"unknown provider", "unknown", false, nil, "unknown provider"},
		{"empty name", "", false, nil, "invalid driver ID format"},
		{"empty version", "nvidia:", false, nil, "invalid driver ID format"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			provider := mocks.NewMockProvider(ctrl)
			provider.EXPECT().GetID().Return("nvidia").AnyTimes()
			provider.EXPECT().GetName().Return("NVIDIA").AnyTimes()
			if tt.query {
				provider.EXPECT().Remove([]api.DriverID{{ProviderID: "nvidia"}}).Return(nil, tt.queryErr)
			}
			deps := api.CoreDeps{
				Providers: []api.Provider{provider}, PackageManager: mocks.NewMockPackageManager(ctrl),
				RepositoryManager: mocks.NewMockRepositoryManager(ctrl),
			}
			err := Execute([]string{"radii", "rm", tt.argument}, deps, "test")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}
