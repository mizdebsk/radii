package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/mocks"
)

func TestInstallHelpAfterKernelOptions(t *testing.T) {
	for _, args := range [][]string{
		{"--kernel-variant", "64k", "--help"},
		{"--kernel", "6.12.0-211.51.1.el10_2", "--help"},
		{"--kernel", "6.12.0-211.51.1.el10_2", "--kernel-variant", "64k", "--help"},
		{"--kernel-variant", "", "--help"},
		{"--kernel-variant", "64k", "-h"},
		{"--kernel-variant=64k", "--help"},
		{"-K", "6.12.0-211.51.1.el10_2", "--help"},
		{"--help"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			stdout := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = stdout }()

			ctrl := gomock.NewController(t)
			deps := api.CoreDeps{
				Providers:         []api.Provider{mocks.NewMockProvider(ctrl)},
				PackageManager:    mocks.NewMockPackageManager(ctrl),
				RepositoryManager: mocks.NewMockRepositoryManager(ctrl),
			}
			if err := Execute(append([]string{"radii", "install"}, args...), deps, "test"); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"Usage:", "radii install", "-K, --kernel RELEASE", "--kernel-variant VARIANT"} {
				if !strings.Contains(string(data), want) {
					t.Errorf("help output missing %q: %s", want, data)
				}
			}
		})
	}
}
