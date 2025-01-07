package comparator_test

import (
	"testing"

	"github.com/ryanjohnsontv/go-homeassistant/shared/state"
	"github.com/ryanjohnsontv/go-homeassistant/shared/utils/comparator"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name         string
		condition    comparator.Condition
		stateValue   state.Value
		expected     bool
		expectsError bool
	}{
		{
			name: "String equals",
			condition: comparator.Condition{
				Type:  comparator.ConditionEquals,
				Value: "123",
			},
			stateValue:   state.Value("123"),
			expected:     true,
			expectsError: false,
		},
		{
			name: "String not equals",
			condition: comparator.Condition{
				Type:  comparator.ConditionNotEquals,
				Value: "456",
			},
			stateValue:   state.Value("123"),
			expected:     true,
			expectsError: false,
		},
		{
			name: "Int equals",
			condition: comparator.Condition{
				Type:  comparator.ConditionEquals,
				Value: 123,
			},
			stateValue:   state.Value("123"),
			expected:     true,
			expectsError: false,
		},
		{
			name: "Float greater",
			condition: comparator.Condition{
				Type:  comparator.ConditionGreater,
				Value: 123.45,
			},
			stateValue:   state.Value("124.00"),
			expected:     true,
			expectsError: false,
		},
		{
			name: "Boolean equals",
			condition: comparator.Condition{
				Type:  comparator.ConditionEquals,
				Value: true,
			},
			stateValue:   state.Value("true"),
			expected:     true,
			expectsError: false,
		},
		{
			name: "Slice contains",
			condition: comparator.Condition{
				Type:  comparator.ConditionIn,
				Value: []int{123, 456, 789},
			},
			stateValue:   state.Value("123"),
			expected:     true,
			expectsError: false,
		},
		{
			name: "Slice does not contain",
			condition: comparator.Condition{
				Type:  comparator.ConditionNotIn,
				Value: []float64{123.45, 678.90},
			},
			stateValue:   state.Value("456.78"),
			expected:     true,
			expectsError: false,
		},
		{
			name: "Unsupported type",
			condition: comparator.Condition{
				Type:  comparator.ConditionEquals,
				Value: map[string]string{"key": "value"},
			},
			stateValue:   state.Value("123"),
			expected:     false,
			expectsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := comparator.Compare(tt.condition, tt.stateValue)

			if tt.expectsError {
				if err == nil {
					t.Fatalf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if result != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}
