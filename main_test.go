package main

import (
	"testing"
)

func TestDecodeResponseFlatValue(t *testing.T) {
	testV := `"value"`
	if v, err := decodeResponse([]byte(testV), nil); err != nil {
		t.Fatalf("failed to decode flat value: %v", err)
	} else if v != "value" {
		t.Fatalf("flat values do not match: %v and %v", v, testV)
	}
}

func TestDecodeNestedField(t *testing.T) {
	testV := `{"field": {"nested": "value"}}`
	if v, err := decodeResponse([]byte(testV), []string{"field", "nested"}); err != nil {
		t.Fatalf("failed to decode flat value: %v", err)
	} else if v != "value" {
		t.Fatalf("flat values do not match: %v and %v", v, testV)
	}
}

func TestDecodeMissingField(t *testing.T) {
	testV := `{"field": {"nested": "value"}}`
	if _, err := decodeResponse([]byte(testV), []string{"field", "missing"}); err == nil {
		t.Fatal("should fail with missing field")
	}
	if _, err := decodeResponse([]byte(testV), []string{"field", "nested", "missing"}); err == nil {
		t.Fatal("should fail with missing field")
	}
}
