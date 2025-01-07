package comparator

import (
	"cmp"
	"fmt"
	"reflect"

	"github.com/ryanjohnsontv/go-homeassistant/shared/state"
)

type (
	ConditionType int

	Condition struct {
		Type  ConditionType
		Value any
	}
)

const (
	ConditionEquals ConditionType = iota
	ConditionNotEquals
	ConditionGreater
	ConditionGreaterEqual
	ConditionLess
	ConditionLessEqual
	ConditionIn
	ConditionNotIn
)

func Compare(condition Condition, state state.Value) (bool, error) {
	switch v := condition.Value.(type) {
	case string:
		return compare(condition.Type, v, state.String())
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uintptr:
		intState, err := state.AsNumber().Int64()
		if err != nil {
			return false, err
		}

		intVal, err := toInt64(v)
		if err != nil {
			return false, err
		}

		return compare(condition.Type, intState, intVal)
	case float32, float64:
		floatState, err := state.AsNumber().Float64()
		if err != nil {
			return false, err
		}

		floatVal, err := toFloat64(v)
		if err != nil {
			return false, err
		}

		return compare(condition.Type, floatState, floatVal)
	case bool:
		boolState, err := state.AsBool()
		if err != nil {
			return false, err
		}

		return compareBool(condition.Type, boolState, v)
	case []string:
		return compareSlice(condition.Type, state.String(), v)
	case []int, []int8, []int16, []int32, []uint, []uint8, []uint16, []uint32, []uint64, []uintptr:
		intState, err := state.AsNumber().Int64()
		if err != nil {
			return false, err
		}

		intSlice, err := toInt64Slice(v)
		if err != nil {
			return false, err
		}

		return compareSlice(condition.Type, intState, intSlice)
	case []int64:
		intState, err := state.AsNumber().Int64()
		if err != nil {
			return false, err
		}

		return compareSlice(condition.Type, intState, v)
	case []float32:
		floatState, err := state.AsNumber().Float64()
		if err != nil {
			return false, err
		}

		floatSlice, err := toFloat64Slice(v)
		if err != nil {
			return false, err
		}

		return compareSlice(condition.Type, floatState, floatSlice)
	case []float64:
		floatState, err := state.AsNumber().Float64()
		if err != nil {
			return false, err
		}

		return compareSlice(condition.Type, floatState, v)
	default:
		return false, fmt.Errorf("unsupported data type: %T", v)
	}
}

func compare[T cmp.Ordered](conditionalType ConditionType, state, value T) (bool, error) {
	switch conditionalType {
	case ConditionEquals:
		return state == value, nil
	case ConditionNotEquals:
		return state != value, nil
	case ConditionGreater:
		return state > value, nil
	case ConditionGreaterEqual:
		return state >= value, nil
	case ConditionLess:
		return state < value, nil
	case ConditionLessEqual:
		return state <= value, nil
	default:
		return false, fmt.Errorf("unupported conditional type for %T: %v", value, conditionalType)
	}
}

func compareBool(conditionalType ConditionType, state, value bool) (bool, error) {
	switch conditionalType {
	case ConditionEquals:
		return state == value, nil
	case ConditionNotEquals:
		return state != value, nil
	default:
		return false, fmt.Errorf("unupported conditional type for %T: %v", value, conditionalType)
	}
}

func compareSlice[T comparable](conditionalType ConditionType, state T, values []T) (bool, error) {
	contains := func(v T) bool {
		for _, val := range values {
			if v == val {
				return true
			}
		}

		return false
	}

	switch conditionalType {
	case ConditionIn:
		return contains(state), nil
	case ConditionNotIn:
		return !contains(state), nil
	default:
		return false, fmt.Errorf("unupported conditional type for %T: %v", values, conditionalType)
	}
}

func toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uintptr:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("type %T is not convertible to int64: %v", v, v)
	}
}

func toFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("type %T is not convertible to float64: %v", v, v)
	}
}

// Converts a generic slice of integers to []int64
func toInt64Slice(slice any) ([]int64, error) {
	value := reflect.ValueOf(slice)
	if value.Kind() != reflect.Slice {
		return nil, fmt.Errorf("value is not a slice: %T", slice)
	}

	var result []int64

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i).Interface()

		intVal, err := toInt64(elem)
		if err != nil {
			return nil, fmt.Errorf("element %d is not convertible to int64: %v", i, elem)
		}

		result = append(result, intVal)
	}

	return result, nil
}

// Converts a generic slice of floats to []float64
func toFloat64Slice(slice any) ([]float64, error) {
	value := reflect.ValueOf(slice)
	if value.Kind() != reflect.Slice {
		return nil, fmt.Errorf("value is not a slice: %T", slice)
	}

	var result []float64

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i).Interface()

		floatVal, err := toFloat64(elem)
		if err != nil {
			return nil, fmt.Errorf("element %d is not convertible to float64: %v", i, elem)
		}

		result = append(result, floatVal)
	}

	return result, nil
}
