package switches

import (
	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
)

const domain = domains.Switch

type Domain struct {
	caller types.ActionCaller
}

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
