package binson

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestEmptyBinsonObject(t *testing.T) {
	var exp = []byte("\x40\x41") // {}
	var b bytes.Buffer
	var e = NewEncoder(&b)

	// ----
	e.Begin()
	e.End()
	// ----

	e.Flush()

	if !bytes.Equal(exp, b.Bytes()) {
		t.Errorf("Binson encoder failure: expected 0x%v", hex.EncodeToString(exp))
	}
}

func TestEmptyBinsonArray(t *testing.T) {
	var exp = []byte("\x42\x43") // []
	var b bytes.Buffer
	var e = NewEncoder(&b)

	// ----
	e.BeginArray()
	e.EndArray()
	// ----

	e.Flush()

	if !bytes.Equal(exp, b.Bytes()) {
		t.Errorf("Binson encoder failure: expected 0x%v", hex.EncodeToString(exp))
	}
}

func TestObjectWithUTF8Name(t *testing.T) {
	var exp = []byte("\x40\x14\x06\xe7\x88\x85\xec\x9b\xa1\x10\x7b\x41") // {"爅웡":123}
	var b bytes.Buffer
	var e = NewEncoder(&b)

	// ----
	e.Begin()
	e.Name("爅웡")
	e.Integer(123)
	e.End()
	// ----

	e.Flush()

	if !bytes.Equal(exp, b.Bytes()) {
		t.Errorf("Binson encoder failure: expected 0x%v", hex.EncodeToString(exp))
	}
}

func TestNestedObjectsWithEmptyKeyNames(t *testing.T) {
	// {"":{"":{"":{}}}}
	var exp = []byte("\x40\x14\x00\x40\x14\x00\x40\x14\x00\x40\x41\x41\x41\x41")
	var b bytes.Buffer
	var e = NewEncoder(&b)

	// ----
	e.Begin()
	e.Name("")
	e.Begin()
	e.Name("")
	e.Begin()
	e.Name("")
	e.Begin()
	e.End()
	e.End()
	e.End()
	e.End()
	// ----

	e.Flush()

	if !bytes.Equal(exp, b.Bytes()) {
		t.Errorf("Binson encoder failure: expected 0x%v", hex.EncodeToString(exp))
	}
}

func TestNestedArraysAsObjectValue(t *testing.T) {
	// {"b":[[[]]]}
	var exp = []byte("\x40\x14\x01\x62\x42\x42\x42\x43\x43\x43\x41")
	var b bytes.Buffer
	var e = NewEncoder(&b)

	// ----
	e.Begin()
	e.Name("b")
	e.BeginArray()
	e.BeginArray()
	e.BeginArray()
	e.EndArray()
	e.EndArray()
	e.EndArray()
	e.End()
	// ----

	e.Flush()

	if !bytes.Equal(exp, b.Bytes()) {
		t.Errorf("Binson encoder failure: expected 0x%v", hex.EncodeToString(exp))
	}
}

func TestNestedStructures1AsObjectValue(t *testing.T) {
	// {"b":[[],{},[]]}
	var exp = []byte("\x40\x14\x01\x62\x42\x42\x43\x40\x41\x42\x43\x43\x41")
	var b bytes.Buffer
	var e = NewEncoder(&b)

	// ----
	e.Begin()
	e.Name("b")
	e.BeginArray()
	e.BeginArray()
	e.EndArray()
	e.Begin()
	e.End()
	e.BeginArray()
	e.EndArray()
	e.EndArray()
	e.End()
	// ----

	e.Flush()

	if !bytes.Equal(exp, b.Bytes()) {
		t.Errorf("Binson encoder failure: expected 0x%v", hex.EncodeToString(exp))
	}
}

func TestNestedStructures2AsObjectValue(t *testing.T) {
	// {"b":[[{}],[{}]]}
	var exp = []byte("\x40\x14\x01\x62\x42\x42\x40\x41\x43\x42\x40\x41\x43\x43\x41")
	var b bytes.Buffer
	var e = NewEncoder(&b)

	// ----
	e.Begin()
	e.Name("b")
	e.BeginArray()
	e.BeginArray()
	e.Begin()
	e.End()
	e.EndArray()
	e.BeginArray()
	e.Begin()
	e.End()
	e.EndArray()
	e.EndArray()
	e.End()
	// ----

	e.Flush()

	if !bytes.Equal(exp, b.Bytes()) {
		t.Errorf("Binson encoder failure: expected 0x%v", hex.EncodeToString(exp))
	}
}

func TestComplexObjectStructure1(t *testing.T) {
	// {"abc":{"cba":{}}, "b":{"abc":{}}}
	var exp = []byte("\x40\x14\x03\x61\x62\x63\x40\x14\x03\x63\x62\x61\x40\x41\x41\x14\x01\x62\x40\x14\x03\x61\x62\x63\x40\x41\x41\x41")
	var b bytes.Buffer
	var e = NewEncoder(&b)

	// ----
	e.Begin()
	e.Name("abc")
	e.Begin()
	e.Name("cba")
	e.Begin()
	e.End()
	e.End()
	e.Name("b")
	e.Begin()
	e.Name("abc")
	e.Begin()
	e.End()
	e.End()
	e.End()
	// ----

	e.Flush()

	if !bytes.Equal(exp, b.Bytes()) {
		t.Errorf("Binson encoder failure: expected 0x%v", hex.EncodeToString(exp))
	}
}

func TestComplexObjectStructure2(t *testing.T) {
	// {"b":[true,13,"cba",{"abc":false, "b":"0x008100ff00", "cba":"abc"},9223372036854775807]}
	var exp = []byte(
		"\x40\x14\x01\x62\x42\x44\x10\x0d\x14\x03\x63\x62\x61\x40\x14\x03" +
			"\x61\x62\x63\x45\x14\x01\x62\x18\x05\x00\x81\x00\xff\x00\x14\x03" +
			"\x63\x62\x61\x14\x03\x61\x62\x63\x41\x13\xff\xff\xff\xff\xff\xff" +
			"\xff\x7f\x43\x41",
	)
	var b bytes.Buffer
	var e = NewEncoder(&b)

	// ----
	e.Begin()
	e.Name("b")
	e.BeginArray()
	e.Bool(true)
	e.Integer(13)
	e.String("cba")
	e.Begin()
	e.Name("abc")
	e.Bool(false)
	e.Name("b")
	e.Bytes([]byte("\x00\x81\x00\xff\x00"))
	e.Name("cba")
	e.String("abc")
	e.End()
	e.Integer(9223372036854775807)
	e.EndArray()
	e.End()
	// ----

	e.Flush()

	if !bytes.Equal(exp, b.Bytes()) {
		t.Errorf("Binson encoder failure: expected 0x%v", hex.EncodeToString(exp))
	}
}
