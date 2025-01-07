package alarmcontrolpanel

type State string

const (
	ArmedAway         State = "armed_away"
	ArmedCustomBypass State = "armed_custom_bypass"
	ArmedHome         State = "armed_home"
	ArmedNight        State = "armed_night"
	ArmedVacation     State = "armed_vacation"
	Disarmed          State = "disarmed"
	Pending           State = "pending"
	Triggered         State = "triggered"
)
