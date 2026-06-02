package opt

import (
	"encoding/json"
	"testing"
)

type testStruct struct {
	Value V[int] `json:"value"`
	Other string `json:"other"`
}

type omitEmptyStruct struct {
	Value V[int] `json:"value,omitzero"`
	Other string `json:"other"`
}

// test unmarshaling of missing value.
func TestUnmarshalJSONMissing(t *testing.T) {
	var ts testStruct
	err := json.Unmarshal([]byte(`{"other": "test"}`), &ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.Value.Exists {
		t.Error("expected Ok to be false, got true")
	}
}

// test unmarshaling of null value.
func TestUnmarshalJSONNull(t *testing.T) {
	var ts testStruct
	err := json.Unmarshal([]byte(`{"other": "test", "value": null}`), &ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.Value.Exists {
		t.Error("expected Ok to be false, got true")
	}
}

// test marshaling of empty value.
func TestMarshalJSONEmpty(t *testing.T) {
	ts := testStruct{
		Other: "test",
	}
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `{"value":null,"other":"test"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

// test marshaling of empty value.
func TestMarshalJSONExisting(t *testing.T) {
	ts := testStruct{
		Value: Val(42),
		Other: "test",
	}
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `{"value":42,"other":"test"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

// test marshaling of a null value.
func TestMarshalJSONOmitZeroExisting(t *testing.T) {
	ts := omitEmptyStruct{
		Value: Val(42),
		Other: "test",
	}
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `{"value":42,"other":"test"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

// test marshaling of a null value.
func TestMarshalJSONOmitZeroMissing(t *testing.T) {
	ts := omitEmptyStruct{
		Other: "test",
	}
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `{"other":"test"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}
