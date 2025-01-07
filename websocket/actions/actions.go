package actions

import (
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
	"github.com/ryanjohnsontv/go-homeassistant/websocket/actions/alarmcontrolpanel"
	"github.com/ryanjohnsontv/go-homeassistant/websocket/actions/button"
	"github.com/ryanjohnsontv/go-homeassistant/websocket/actions/climate"
	"github.com/ryanjohnsontv/go-homeassistant/websocket/actions/light"
	"github.com/ryanjohnsontv/go-homeassistant/websocket/actions/switches"
)

type (
	Actions struct {
		AlarmControlPanel *alarmcontrolpanel.Domain
		Button            *button.Domain
		Climate           *climate.Domain
		Light             *light.Domain
		Switch            *switches.Domain
	}
)

func NewActionService(caller types.ActionCaller) *Actions {
	return &Actions{
		AlarmControlPanel: alarmcontrolpanel.NewDomain(caller),
		Button:            button.NewDomain(caller),
		Climate:           climate.NewDomain(caller),
		Light:             light.NewDomain(caller),
		Switch:            switches.NewDomain(caller),
	}
}
