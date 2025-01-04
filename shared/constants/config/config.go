package config

import "encoding/json"

type HassConfigState string

const (
	StateNotRunning HassConfigState = "NOT_RUNNING"
	StateStarting   HassConfigState = "STARTING"
	StateRunning    HassConfigState = "RUNNING"
	StateStopping   HassConfigState = "STOPPING"
	StateFinalWrite HassConfigState = "FINAL_WRITE"
)

func (s *HassConfigState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	*s = HassConfigState(str)

	return nil
}
