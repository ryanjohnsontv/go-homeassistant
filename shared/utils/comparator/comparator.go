package comparator

import (
	"cmp"
	"fmt"
	"reflect"
	"strconv"

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

func Compare(condition Condition, state any) (bool, error) {
	switch v := condition.Value.(type) {
	case string:
		stringState, err := ToString(state)
		if err != nil {
			return false, err
		}

		return compare(condition.Type, v, stringState)
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uintptr:
		intState, err := ToInt64(state)
		if err != nil {
			return false, err
		}

		intVal, err := ToInt64(v)
		if err != nil {
			return false, err
		}

		return compare(condition.Type, intState, intVal)
	case float32, float64:
		floatState, err := ToFloat64(state)
		if err != nil {
			return false, err
		}

		floatVal, err := ToFloat64(v)
		if err != nil {
			return false, err
		}

		return compare(condition.Type, floatState, floatVal)
	case bool:
		boolState, err := ToBool(state)
		if err != nil {
			return false, err
		}

		return compareBool(condition.Type, boolState, v)
	case []string:
		stringState, err := ToString(state)
		if err != nil {
			return false, err
		}

		return compareSlice(condition.Type, stringState, v)
	case []int, []int8, []int16, []int32, []uint, []uint8, []uint16, []uint32, []uint64, []uintptr:
		intState, err := ToInt64(state)
		if err != nil {
			return false, err
		}

		intSlice, err := ToInt64Slice(v)
		if err != nil {
			return false, err
		}

		return compareSlice(condition.Type, intState, intSlice)
	case []int64:
		intState, err := ToInt64(state)
		if err != nil {
			return false, err
		}

		return compareSlice(condition.Type, intState, v)
	case []float32:
		floatState, err := ToFloat64(state)
		if err != nil {
			return false, err
		}

		floatSlice, err := ToFloat64Slice(v)
		if err != nil {
			return false, err
		}

		return compareSlice(condition.Type, floatState, floatSlice)
	case []float64:
		floatState, err := ToFloat64(state)
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

type Bool interface {
	Bool() (bool, error)
}

func ToBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case state.Value:
		return v.Bool()
	case string:
		return state.StringToBool(v)
	case Bool:
		return v.Bool()
	default:
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.String {
			return state.StringToBool(val.String())
		}

		return false, fmt.Errorf("type %T is not convertible to bool", value)
	}
}

type Int interface {
	Int64() (int64, error)
}

func ToInt64(value any) (int64, error) {
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
	case state.Value:
		return v.Int64()
	case string:
		return parseInt64(v)
	case Int:
		return v.Int64()
	default:
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.String {
			return parseInt64(val.String())
		}

		return 0, fmt.Errorf("type %T is not convertible to int64: %v", v, v)
	}
}

func parseInt64(s string) (int64, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

type Float interface {
	Float64() (float64, error)
}

func ToFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case state.Value:
		return v.Float64()
	case string:
		return parseFloat64(v)
	case Float:
		return v.Float64()
	default:
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.String {
			return parseFloat64(val.String())
		}

		return 0, fmt.Errorf("type %T is not convertible to float64: %v", v, v)
	}
}

func parseFloat64(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return f, nil
}

func ToString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case state.Value:
		return v.String(), nil
	default:
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.String {
			return val.String(), nil
		}

		return "", fmt.Errorf("type %T is not convertible to int64: %v", value, value)
	}
}

// Converts a generic slice of integers to []int64
func ToInt64Slice(slice any) ([]int64, error) {
	value := reflect.ValueOf(slice)
	if value.Kind() != reflect.Slice {
		return nil, fmt.Errorf("value is not a slice: %T", slice)
	}

	var result []int64

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i).Interface()

		intVal, err := ToInt64(elem)
		if err != nil {
			return nil, fmt.Errorf("element %d is not convertible to int64: %v", i, elem)
		}

		result = append(result, intVal)
	}

	return result, nil
}

// Converts a generic slice of floats to []float64
func ToFloat64Slice(slice any) ([]float64, error) {
	value := reflect.ValueOf(slice)
	if value.Kind() != reflect.Slice {
		return nil, fmt.Errorf("value is not a slice: %T", slice)
	}

	var result []float64

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i).Interface()

		floatVal, err := ToFloat64(elem)
		if err != nil {
			return nil, fmt.Errorf("element %d is not convertible to float64: %v", i, elem)
		}

		result = append(result, floatVal)
	}

	return result, nil
}
