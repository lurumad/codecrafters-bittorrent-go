package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestErrBencodeString(t *testing.T) {
	bencodedString := "5hello"

	bencodeDecoded := decodeBencode(bencodedString)

	if !errors.Is(bencodeDecoded.err, ErrBencodeString) {
		t.Errorf("expected ErrBencodeString - got: %v", bencodeDecoded.err)
	}

	if bencodeDecoded.value != "" {
		t.Errorf("error result should be empty - got: %v", bencodeDecoded.value)
	}

	if bencodeDecoded.end > 0 {
		t.Errorf("error end of string should be 0 - got %d", bencodeDecoded.end)
	}
}

func TestErrDecodeBencodeInteger(t *testing.T) {
	bencodedString := "i52"

	bencodeDecoded := decodeBencode(bencodedString)

	if !errors.Is(bencodeDecoded.err, ErrBencodeInteger) {
		t.Errorf("expected ErrBencodeInteger - got: %v", bencodeDecoded.err)
	}

	if bencodeDecoded.value != 0 {
		t.Errorf("error result should be 0 - got: %v", bencodeDecoded.value)
	}

	if bencodeDecoded.end > 0 {
		t.Errorf("error end should be 0 - got %d", bencodeDecoded.end)
	}
}

func TestDecodeBencode(t *testing.T) {
	type testCase struct {
		bencoded string
		want     BencodeType
		end      int
	}

	for _, tc := range []testCase{
		{bencoded: "4:pear", want: "pear", end: 6},
		{bencoded: "5:hello", want: "hello", end: 7},
		{bencoded: "i52e", want: 52, end: 4},
		{bencoded: "l5:helloi52ee", want: []interface{}{"hello", 52}, end: 13},
		{bencoded: "le", want: []interface{}{}, end: 2},
		{bencoded: "lli940e5:appleee", want: []interface{}{[]interface{}{940, "apple"}}, end: 16},
		{bencoded: "lli4eei5ee", want: []interface{}{[]interface{}{4}, 5}, end: 10},
		{bencoded: "d3:foo3:bar5:helloi52ee", want: map[interface{}]interface{}{
			"foo":   "bar",
			"hello": 52,
		}, end: 23},
		{bencoded: "de", want: map[interface{}]interface{}{}, end: 2},
	} {
		bencodeDecoded := decodeBencode(tc.bencoded)

		if bencodeDecoded.err != nil {
			t.Fatal(bencodeDecoded.err)
		}

		if bencodeDecoded.value == "" {
			t.Errorf("%v error result should be empty - got: %v", tc.bencoded, bencodeDecoded.value)
		}

		if !equals(bencodeDecoded.value, tc.want) {
			t.Errorf("%v bad result - want %v, got %v", tc.bencoded, tc.want, bencodeDecoded.value)
		}

		if bencodeDecoded.end != tc.end {
			t.Errorf("%v bad end - want %v, got %v", tc.bencoded, tc.end, bencodeDecoded.end)
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
		return reflect.DeepEqual(a, b)
	case map[interface{}]interface{}:
		b := b.(map[interface{}]interface{})
		if len(a) != len(b) {
			return false
		}
		return reflect.DeepEqual(a, b)
	default:
		return false
	}
}
