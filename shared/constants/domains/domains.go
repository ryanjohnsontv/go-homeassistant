package domains

type Domain string

const (
	AirQuality        Domain = "air_quality"         // https://www.home-assistant.io/integrations/air_quality
	AlarmControlPanel Domain = "alarm_control_panel" // https://www.home-assistant.io/integrations/alarm_control_panel
	AssistSatellite   Domain = "assist_satellite"    // https://www.home-assistant.io/integrations/assist_satellite
	Automation        Domain = "automation"          // https://www.home-assistant.io/docs/automation
	Backup            Domain = "backup"              // https://www.home-assistant.io/integrations/backup
	BinarySensor      Domain = "binary_sensor"       // https://www.home-assistant.io/integrations/binary_sensor
	Button            Domain = "button"              // https://www.home-assistant.io/integrations/button
	Calendar          Domain = "calendar"            // https://www.home-assistant.io/integrations/calendar
	Camera            Domain = "camera"              // https://www.home-assistant.io/integrations/camera
	Climate           Domain = "climate"             // https://www.home-assistant.io/integrations/climate
	Conversation      Domain = "conversation"        // https://www.home-assistant.io/integrations/conversation/
	Cover             Domain = "cover"               // https://www.home-assistant.io/integrations/cover
	Date              Domain = "date"                // https://www.home-assistant.io/integrations/date
	DateTime          Domain = "datetime"            // https://www.home-assistant.io/integrations/datetime
	DeviceTracker     Domain = "device_tracker"      // https://www.home-assistant.io/integrations/device_tracker
	Event             Domain = "event"               // https://www.home-assistant.io/integrations/event
	Fan               Domain = "fan"                 // https://www.home-assistant.io/integrations/fan
	Geolocation       Domain = "geo_location"        // https://www.home-assistant.io/integrations/geo_location
	Group             Domain = "group"               // https://www.home-assistant.io/integrations/group
	Humidifier        Domain = "humidifier"          // https://www.home-assistant.io/integrations/humidifier
	Image             Domain = "image"               // https://www.home-assistant.io/integrations/image
	ImageProcessing   Domain = "image_processing"    // https://www.home-assistant.io/integrations/image_processing
	InputBoolean      Domain = "input_boolean"       // https://www.home-assistant.io/integrations/input_boolean
	InputButton       Domain = "input_button"        // https://www.home-assistant.io/integrations/input_button
	InputDatetime     Domain = "input_datetime"      // https://www.home-assistant.io/integrations/input_datetime
	InputNumber       Domain = "input_number"        // https://www.home-assistant.io/integrations/input_number
	InputSelect       Domain = "input_select"        // https://www.home-assistant.io/integrations/input_select
	InputText         Domain = "input_text"          // https://www.home-assistant.io/integrations/input_text
	LawnMower         Domain = "lawn_mower"          // https://www.home-assistant.io/integrations/lawn_mower
	Light             Domain = "light"               // https://www.home-assistant.io/integrations/light
	Lock              Domain = "lock"                // https://www.home-assistant.io/integrations/lock
	MediaPlayer       Domain = "media_player"        // https://www.home-assistant.io/integrations/media_player
	Notifications     Domain = "notify"              // https://www.home-assistant.io/integrations/notify
	Number            Domain = "number"              // https://www.home-assistant.io/integrations/number
	Person            Domain = "person"              // https://www.home-assistant.io/integrations/person
	Remote            Domain = "remote"              // https://www.home-assistant.io/integrations/remote
	Scene             Domain = "scene"               // https://www.home-assistant.io/integrations/scene
	Script            Domain = "script"              // https://www.home-assistant.io/integrations/script
	Select            Domain = "select"              // https://www.home-assistant.io/integrations/select
	Sensor            Domain = "sensor"              // https://www.home-assistant.io/integrations/sensor
	Siren             Domain = "siren"               // https://www.home-assistant.io/integrations/siren
	STT               Domain = "stt"                 // https://www.home-assistant.io/integrations/stt
	Sun               Domain = "sun"                 // https://www.home-assistant.io/integrations/sun
	Switch            Domain = "switch"              // https://www.home-assistant.io/integrations/switch
	TagScanned        Domain = "tag_scanned"         // https://www.home-assistant.io/integrations/tag
	Text              Domain = "text"                // https://www.home-assistant.io/integrations/text
	Time              Domain = "time"                // https://www.home-assistant.io/integrations/time
	Todo              Domain = "todo"                // https://www.home-assistant.io/integrations/todo
	TTS               Domain = "tts"                 // https://www.home-assistant.io/integrations/tts
	Update            Domain = "update"              // https://www.home-assistant.io/integrations/update
	Vacuum            Domain = "vacuum"              // https://www.home-assistant.io/integrations/vacuum
	Valve             Domain = "valve"               // https://www.home-assistant.io/integrations/valve
	WakeWord          Domain = "wake_word"           // https://www.home-assistant.io/integrations/wake_word
	WaterHeater       Domain = "water_heater"        // https://www.home-assistant.io/integrations/water_heater
	Weather           Domain = "weather"             // https://www.home-assistant.io/integrations/weather
	Zone              Domain = "zone"                // https://www.home-assistant.io/integrations/zone
)

func (d *Domain) String() string {
	return string(*d)
}
