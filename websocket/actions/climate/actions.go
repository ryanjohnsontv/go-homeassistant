package climate

import (
	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
)

const domain = domains.Climate

type (
	Domain struct {
		caller types.ActionCaller
	}

	setAuxHeatData struct {
		AuxHeat bool `json:"aux_heat"`
	}

	setFanModeData struct {
		Mode string `json:"fan_mode"`
	}

	setHVACModeData struct {
		Mode string `json:"hvac_mode"`
	}

	setHumidityData struct {
		Humidity float32 `json:"humidity"`
	}

	setPresetModeData struct {
		Mode string `json:"preset_mode"`
	}

	setSwingModeData struct {
		Mode string `json:"swing_mode"`
	}

	setSwingHorizontalModeData struct {
		Mode string `json:"swing_horizontal_mode"`
	}

	SetTemperatureData struct {
		Temperature    float32 `json:"temperature,omitempty"`
		TargetTempHigh float32 `json:"target_temp_high,omitempty"`
		TargetTempLow  float32 `json:"target_temp_low,omitempty"`
		HVACMode       string  `json:"hvac_mode,omitempty"`
	}
)

func NewDomain(c types.ActionCaller) *Domain {
	return &Domain{caller: c}
}

// SetAuxHeat creates a ServiceTargetBuilder for the set_aux_heat service.
func (d *Domain) SetAuxHeat(enabled bool) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "set_aux_heat",
		Domain: domain,
		Data: setAuxHeatData{
			AuxHeat: enabled,
		},
	}
}

// SetFanMode creates a ServiceTargetBuilder for the set_fan_mode service.
func (d *Domain) SetFanMode(mode string) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "set_fan_mode",
		Domain: domain,
		Data: setFanModeData{
			Mode: mode,
		},
	}
}

// SetHVACMode creates a ServiceTargetBuilder for the set_hvac_mode service.
func (d *Domain) SetHVACMode(mode string) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "set_hvac_mode",
		Domain: domain,
		Data: setHVACModeData{
			Mode: mode,
		},
	}
}

// SetHumidity creates a ServiceTargetBuilder for the set_humidity service.
func (d *Domain) SetHumidity(humidity float32) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "set_humidity",
		Domain: domain,
		Data: setHumidityData{
			Humidity: humidity,
		},
	}
}

// SetPresetMode creates a ServiceTargetBuilder for the set_preset_mode service.
func (d *Domain) SetPresetMode(mode string) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "set_preset_mode",
		Domain: domain,
		Data: setPresetModeData{
			Mode: mode,
		},
	}
}

// SetSwingMode creates a ServiceTargetBuilder for the set_swing_mode service.
func (d *Domain) SetSwingMode(mode string) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "set_swing_mode",
		Domain: domain,
		Data: setSwingModeData{
			Mode: mode,
		},
	}
}

// SetSwingHorizontalMode creates a ServiceTargetBuilder for the set_swing_horizontal_mode service.
func (d *Domain) SetSwingHorizontalMode(mode string) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "set_swing_mode",
		Domain: domain,
		Data: setSwingHorizontalModeData{
			Mode: mode,
		},
	}
}

// SetTemperature creates a ServiceTargetBuilder for the set_temperature service.
func (d *Domain) SetTemperature() *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "set_temperature",
		Domain: domain,
	}
}

// TurnOn creates a ServiceTargetBuilder for the turn_on service.
func (d *Domain) TurnOn() *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "turn_on",
		Domain: domain,
	}
}

// TurnOff creates a ServiceTargetBuilder for the turn_off service.
func (d *Domain) TurnOff() *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "turn_off",
		Domain: domain,
	}
}

// Toggle creates a ServiceTargetBuilder for the toggle service.
func (d *Domain) Toggle() *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "toggle",
		Domain: domain,
	}
}
