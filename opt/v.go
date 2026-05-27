// Package opt provides a generic type for representing a value that may or may not be present.
package opt

import (
	"bytes"
	"encoding/json"
)

// V represents a value that is optional - it may or may not be present. This
// type should be used to represent fields that are optional in a struct or
// function argument. Pointers should never be used to represent optional
// values. Pointers should only be used to avoid copying of large values.
type V[T any] struct {
	Val    T
	Exists bool
}

// Val creates a new V with the given value and ok == true.
func Val[T any](val T) V[T] {
	return V[T]{Val: val, Exists: true}
}

// UnmarshalJSON implements the json.Unmarshaler interface for V. If there is a
// null value, the value is set to its zero value and Exists is set to false.
// Otherwise the value is unmarshaled to Val and Exists is set to true.
func (v *V[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, null) {
		return nil
	}
	var defaultVal T
	err := json.Unmarshal(data, &defaultVal)
	if err != nil {
		return err
	}
	v.Val = defaultVal
	v.Exists = true
	return nil
}

// MarshalJSON implements the json.Marshaler interface for V. If Exists is false, it
// returns null, otherwise it returns the marshaled value.
func (v V[T]) MarshalJSON() ([]byte, error) {
	if !v.Exists {
		return null, nil
	}
	return json.Marshal(v.Val)
}

var null = []byte("null")
