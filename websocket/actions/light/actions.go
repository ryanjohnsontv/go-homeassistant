package light

import (
	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
)

const domain = domains.Light

type (
	Domain struct {
		caller types.ActionCaller
	}

	TurnOnData struct {
		Transition        int       `json:"transition,omitempty"`
		Profile           string    `json:"profile,omitempty"`
		HSColor           []float32 `json:"hs_color,omitempty"`
		XYColor           []float32 `json:"xy_color,omitempty"`
		RGBColor          []int     `json:"rgb,omitempty"`
		RGBWColor         []int     `json:"rgbw,omitempty"`
		RGBWWColor        []int     `json:"rgbww,omitempty"`
		ColorTempKelvin   int       `json:"color_temp_kelvin,omitempty"`
		ColorName         string    `json:"color_name,omitempty"`
		Brightness        int       `json:"brightness,omitempty"`
		BrightnessPct     int       `json:"brightness_pct,omitempty"`
		BrightnessStep    int       `json:"brightness_step,omitempty"`
		BrightnessStepPct int       `json:"brightness_step_pct,omitempty"`
		White             bool      `json:"white,omitempty"`
		Flash             Flash     `json:"flash,omitempty"`
		Effect            string    `json:"effect,omitempty"`
	}

	TurnOffData struct {
		Transition int   `json:"transition,omitempty"`
		Flash      Flash `json:"flash,omitempty"`
	}

	Flash string
)

const (
	FlashShort Flash = "short"
	FlashLong  Flash = "long"
)

func NewDomain(c types.ActionCaller) *Domain {
	return &Domain{caller: c}
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
