package intel

import "github.com/mizdebsk/radii/internal/hwdetect"

const intelVendor = "8086"

func detectHardware(modaliasRoot string) (bool, error) {
	return hwdetect.ScanPCIModaliases(modaliasRoot, "Intel GPU or accelerator", isCompatibleIntelHardware)
}

func isCompatibleIntelHardware(modalias hwdetect.PCIModalias) bool {
	return modalias.VendorID == intelVendor &&
		(modalias.BaseClass == hwdetect.PCIClassDisplay ||
			modalias.BaseClass == hwdetect.PCIClassProcessingAccelerator)
}
