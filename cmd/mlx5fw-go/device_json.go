package main

import (
	"github.com/Civil/mlx5fw-go/pkg/interfaces"
)

// DeviceQueryJSON extends the standard query JSON with device-only extras (e.g., MCQI)
type DeviceQueryJSON struct {
	*QueryJSONOutput
	MCQI map[string]any `json:"mcqi,omitempty"`
}

// outputDeviceJSON merges base info with extra (e.g., mcqi activation) and prints JSON
func outputDeviceJSON(info *interfaces.FirmwareInfo, extra map[string]any) error {
	base := convertToQueryJSON(info)
	out := DeviceQueryJSON{QueryJSONOutput: base}
	// Add device specific extras
	if extra != nil {
		mcqi := map[string]any{}
		if v, ok := extra["ActivationMethod"]; ok {
			mcqi["activation_method"] = v
		}
		if v, ok := extra["MCQI_VersionString"]; ok {
			mcqi["version_string"] = v
		}
		if v, ok := extra["MCQI_SupportedInfoMask"]; ok {
			mcqi["supported_info_mask"] = v
		}
		if v, ok := extra["MCQI_ComponentSize"]; ok {
			mcqi["component_size"] = v
		}
		if v, ok := extra["MCQI_MaxComponentSize"]; ok {
			mcqi["max_component_size"] = v
		}
		// Header fields for transparency
		if v, ok := extra["MCQI_Activation_InfoType"]; ok {
			mcqi["activation_info_type"] = v
		}
		if v, ok := extra["MCQI_Activation_InfoSize"]; ok {
			mcqi["activation_info_size"] = v
		}
		if v, ok := extra["MCQI_Activation_DataSize"]; ok {
			mcqi["activation_data_size"] = v
		}
		if v, ok := extra["MCQI_Version_InfoType"]; ok {
			mcqi["version_info_type"] = v
		}
		if v, ok := extra["MCQI_Version_InfoSize"]; ok {
			mcqi["version_info_size"] = v
		}
		if v, ok := extra["MCQI_Version_DataSize"]; ok {
			mcqi["version_data_size"] = v
		}
		if v, ok := extra["MCQI_Cap_InfoType"]; ok {
			mcqi["capabilities_info_type"] = v
		}
		if v, ok := extra["MCQI_Cap_InfoSize"]; ok {
			mcqi["capabilities_info_size"] = v
		}
		if v, ok := extra["MCQI_Cap_DataSize"]; ok {
			mcqi["capabilities_data_size"] = v
		}
		if len(mcqi) > 0 {
			out.MCQI = mcqi
		}
	}
	return outputJSON(out)
}
