package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestErrBencodeString(t *testing.T) {
	bencodedString := "5hello"

	got, end, err := decodeBencode(bencodedString)

	if !errors.Is(err, ErrBencodeString) {
		t.Errorf("expected ErrBencodeString - got: %v", err)
	}

	if got != "" {
		t.Errorf("error result should be empty - got: %v", got)
	}

	if end > 0 {
		t.Errorf("error end of string should be 0 - got %d", end)
	}
}

func TestErrDecodeBencodeInteger(t *testing.T) {
	bencodedString := "i52"

	got, end, err := decodeBencode(bencodedString)

	if !errors.Is(err, ErrBencodeInteger) {
		t.Errorf("expected ErrBencodeInteger - got: %v", err)
	}

	if got != 0 {
		t.Errorf("error result should be 0 - got: %v", got)
	}

	if end > 0 {
		t.Errorf("error end should be 0 - got %d", end)
	}
}

func TestDecodeBencode(t *testing.T) {
	type testCase struct {
		bencoded string
		want     BencodeType
		end      int
	}

	for _, tc := range []testCase{
		{bencoded: "5:hello", want: "hello"},
		{bencoded: "i52e", want: 52},
		{bencoded: "l5:helloi52ee", want: []interface{}{"hello", 52}},
		{bencoded: "le", want: []interface{}{}},
	} {
		got, end, err := decodeBencode(tc.bencoded)

		if err != nil {
			t.Fatal(err)
		}

		if got == "" {
			t.Errorf("%v error result should be empty - got: %v", tc.bencoded, got)
		}

		if !equals(got, tc.want) {
			t.Errorf("%v bad result - want %v, got %v", tc.bencoded, tc.want, got)
		}

		if end != len(tc.bencoded) {
			t.Errorf("%v bad end - want %v, got %v", tc.bencoded, len(tc.bencoded), end)
		}
	}
}

func equals(a, b interface{}) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	switch a := a.(type) {
	case string, int:
		return a == b
	case []interface{}:
		b := b.([]interface{})
		if len(a) != len(b) {
			return false
		}
		for index, item := range a {
			if item != b[index] {
				return false
			}
		}
		return true
	default:
		return true
	}
}
