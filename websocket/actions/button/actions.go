package button

import (
	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
)

const domain = domains.Button

type (
	Domain struct {
		caller types.ActionCaller
	}
)

func NewDomain(c types.ActionCaller) *Domain {
	return &Domain{caller: c}
}

// Press creates a ServiceTargetBuilder for the press service.
func (d *Domain) Press() *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "press",
		Domain: domain,
	}
}
