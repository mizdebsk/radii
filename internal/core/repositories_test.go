package core

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
	"github.com/mizdebsk/radii/internal/provider/amd"
	"github.com/mizdebsk/radii/internal/provider/nvidia"
)

func TestProviderChannels(t *testing.T) {
	tests := []struct {
		name      string
		providers []api.Provider
		want      []string
	}{
		{"AMD", []api.Provider{amd.NewProvider(nil)}, []string{"BaseOS", "AppStream", "Extensions"}},
		{"NVIDIA", []api.Provider{nvidia.NewProvider(nil)}, []string{"BaseOS", "AppStream", "Extensions", "Supplementary"}},
		{"Both", []api.Provider{amd.NewProvider(nil), nvidia.NewProvider(nil)}, []string{"BaseOS", "AppStream", "Extensions", "Supplementary"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requiredChannels(tt.providers); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("channels = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListSelectsProviderChannels(t *testing.T) {
	for _, compatibleOnly := range []bool{false, true} {
		name := "All"
		if compatibleOnly {
			name = "CompatibleOnly"
		}
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			rm := mocks.NewMockRepositoryManager(ctrl)
			pm := mocks.NewMockPackageManager(ctrl)
			nv := mocks.NewMockProvider(ctrl)
			am := mocks.NewMockProvider(ctrl)
			third := mocks.NewMockProvider(ctrl)
			providers := []api.Provider{nv, am, third}
			var detections []*gomock.Call
			for i, p := range []*mocks.MockProvider{nv, am, third} {
				p.EXPECT().GetName().Return([]string{"NVIDIA", "AMD GPU", "Third GPU"}[i]).AnyTimes()
				detections = append(detections, p.EXPECT().DetectHardware().Return(i != 0, nil))
			}
			wantChannels := []string{"BaseOS", "Supplementary", "Extensions", "ThirdChannel"}
			wantCount := 3
			if compatibleOnly {
				wantChannels = []string{"BaseOS", "Extensions", "ThirdChannel"}
				wantCount = 2
			} else {
				nv.EXPECT().RequiredChannels().Return([]string{"BaseOS", "Supplementary"})
			}
			am.EXPECT().RequiredChannels().Return([]string{"BaseOS", "Extensions"})
			third.EXPECT().RequiredChannels().Return([]string{"BaseOS", "ThirdChannel"})
			prepared := rm.EXPECT().GetRepoIDs(wantChannels).Return([]string{"selected-repos"}, nil)
			configured := pm.EXPECT().SetEnableRepos([]string{"selected-repos"}).After(prepared)
			for _, detected := range detections {
				prepared.After(detected)
			}
			for i, p := range []*mocks.MockProvider{nv, am, third} {
				if compatibleOnly && i == 0 {
					continue
				}
				id := []string{"nvidia", "amdgpu", "third"}[i]
				p.EXPECT().GetID().Return(id).AnyTimes()
				p.EXPECT().ListAvailable().After(configured).Return([]api.DriverID{{ProviderID: id, Version: "1"}}, nil)
			}
			got, err := List(api.CoreDeps{RepositoryManager: rm, PackageManager: pm, Providers: providers}, false, true, true, compatibleOnly)
			if err != nil || len(got) != wantCount {
				t.Fatalf("List() = %v, %v; want %d drivers", got, err, wantCount)
			}
		})
	}
}
