package intel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectHardware(t *testing.T) {
	// These modaliases are constructed from known PCI IDs, not captured from sysfs.
	for _, tt := range []struct {
		name     string
		modalias string
		want     bool
	}{
		{"Intel VGA controller", "pci:v00008086d000046A6sv00008086sd00000000bc03sc00i00", true},
		{"Intel Arc A770", "pci:v00008086d000056A0sv00008086sd00000000bc03sc00i00", true},
		{"Intel network controller", "pci:v00008086d000015F3sv00008086sd00000000bc02sc00i00", false},
		{"Intel bridge", "pci:v00008086d0000460Dsv00008086sd00000000bc06sc04i00", false},
		{"Intel NPU", "pci:v00008086d00007D1Dsv00008086sd00000000bc12sc00i00", true},
		{"AMD GPU", "pci:v00001002d0000740Fsv00001002sd00000000bc03sc00i00", false},
		{"AMD accelerator", "pci:v00001002d000074A1sv00001002sd000074A1bc12sc00i00", false},
		{"NVIDIA GPU", "pci:v000010DEd000020B0sv000010DEsd00000000bc03sc02i00", false},
		{"invalid modalias", "not-a-pci-modalias", false},
		{"empty sysfs", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if tt.modalias != "" {
				path := filepath.Join(root, "pci0000:00", "0000:00:02.0", "modalias")
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(tt.modalias+"\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			found, err := detectHardware(root)
			if err != nil || found != tt.want {
				t.Fatalf("detectHardware() = %v, %v, want %v, nil", found, err, tt.want)
			}
		})
	}
	if found, err := detectHardware(filepath.Join(t.TempDir(), "missing")); err == nil || found {
		t.Fatalf("detectHardware(missing root) = %v, %v, want false and an error", found, err)
	}
}
