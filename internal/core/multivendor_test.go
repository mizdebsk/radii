package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestAutoDetectRejectsMultipleProvidersBeforePreparation(t *testing.T) {
	for _, count := range []int{2, 3} {
		for _, batch := range []bool{false, true} {
			for _, dryRun := range []bool{false, true} {
				t.Run(fmt.Sprintf("providers=%d/batch=%t/dry=%t", count, batch, dryRun), func(t *testing.T) {
					ctrl := gomock.NewController(t)
					deps := api.CoreDeps{RepositoryManager: mocks.NewMockRepositoryManager(ctrl), PackageManager: mocks.NewMockPackageManager(ctrl)}
					ids := []string{"nvidia", "amdgpu", "third"}[:count]
					for _, id := range ids {
						p := mocks.NewMockProvider(ctrl)
						p.EXPECT().GetID().Return(id).AnyTimes()
						p.EXPECT().GetName().Return(id + " hardware").AnyTimes()
						p.EXPECT().DetectHardware().Return(true, nil)
						deps.Providers = append(deps.Providers, p)
					}
					err := InstallAutoDetect(deps, batch, dryRun, false)
					if err == nil {
						t.Fatal("expected ambiguous auto-detection to fail")
					}
					for _, fragment := range append(ids, "radii install <vendor>:<version>", "radii list --compatible", "radii install --auto-detect --force") {
						if !strings.Contains(err.Error(), fragment) {
							t.Errorf("error %q does not contain %q", err, fragment)
						}
					}
				})
			}
		}
	}
}
