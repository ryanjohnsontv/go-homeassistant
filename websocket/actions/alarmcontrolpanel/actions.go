package alarmcontrolpanel

import (
	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
)

const domain = domains.AlarmControlPanel

type (
	Domain struct {
		caller types.ActionCaller
	}

	alarmArmData struct {
		Code int `json:"code"`
	}
)

func NewDomain(c types.ActionCaller) *Domain {
	return &Domain{caller: c}
}

// ArmHome creates a ServiceTargetBuilder for the alarm_arm_home service.
func (d *Domain) ArmHome(code ...int) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "alarm_arm_home",
		Domain: domain,
		Data: alarmArmData{
			Code: code[0],
		},
	}
}

// ArmAway creates a ServiceTargetBuilder for the alarm_arm_away service.
func (d *Domain) ArmAway(code ...int) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "alarm_arm_away",
		Domain: domain,
		Data: alarmArmData{
			Code: code[0],
		},
	}
}

// ArmNight creates a ServiceTargetBuilder for the alarm_arm_night service.
func (d *Domain) ArmNight(code ...int) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "alarm_arm_night",
		Domain: domain,
		Data: alarmArmData{
			Code: code[0],
		},
	}
}

// ArmVacation creates a ServiceTargetBuilder for the alarm_arm_vacation service.
func (d *Domain) ArmVacation(code ...int) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "alarm_arm_vacation",
		Domain: domain,
		Data: alarmArmData{
			Code: code[0],
		},
	}
}

// Disarm creates a ServiceTargetBuilder for the alarm_disarm service.
func (d *Domain) Disarm(code ...int) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "alarm_disarm",
		Domain: domain,
		Data: alarmArmData{
			Code: code[0],
		},
	}
}

// Trigger creates a ServiceTargetBuilder for the alarm_trigger service.
func (d *Domain) Trigger(code ...int) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "alarm_trigger",
		Domain: domain,
		Data: alarmArmData{
			Code: code[0],
		},
	}
}

// ArmCustomBypass creates a ServiceTargetBuilder for the alarm_arm_custom_bypass service.
func (d *Domain) ArmCustomBypass(code ...int) *types.ServiceTargetBuilder {
	return &types.ServiceTargetBuilder{
		Caller: d.caller,
		Action: "alarm_arm_custom_bypass",
		Domain: domain,
		Data: alarmArmData{
			Code: code[0],
		},
	}
}
