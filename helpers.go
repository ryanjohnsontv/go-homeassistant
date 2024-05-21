package homeassistant

func boolPointer(b bool) *bool {
	return &b
}

//	func ToBool(c commonState) (bool, error) {
//		str, ok := c.value.(string)
//		if !ok {
//			return false, errors.New("state is not a string")
//		}
//		return strconv.ParseBool(str)
//	}
// func structToMap(v interface{}) map[string]interface{} {
// 	// Marshal the struct to JSON
// 	data, err := json.Marshal(v)
// 	if err != nil {
// 		return nil
// 	}

// 	// Unmarshal the JSON into a map[string]interface{}
// 	var result map[string]interface{}
// 	if err := json.Unmarshal(data, &result); err != nil {
// 		return nil
// 	}

// 	// Convert []byte fields to strings
// 	for key, value := range result {
// 		if byteArray, ok := value.([]byte); ok {
// 			result[key] = string(byteArray)
// 		}
// 	}

// 	return result
// }
