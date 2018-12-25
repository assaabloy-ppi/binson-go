package binson

import (
	"bytes"
	"encoding/hex"
	"math"
	"testing"
)

// Binson INTEGER internal representation test data table
var intTable = []struct {
	val int64
	raw []byte
}{
	// int8
	{0, []byte("\x10\x00")},
	{-1, []byte("\x10\xff")},
	{math.MaxInt8, []byte("\x10\x7f")},
	{math.MaxInt8 + 1, []byte("\x11\x80\x00")},
	{math.MinInt8, []byte("\x10\x80")},
	{math.MinInt8 - 1, []byte("\x11\x7f\xff")},

	// int16
	{math.MaxInt16, []byte("\x11\xff\x7f")},
	{math.MaxInt16 + 1, []byte("\x12\x00\x80\x00\x00")},
	{math.MinInt16, []byte("\x11\x00\x80")},
	{math.MinInt16 - 1, []byte("\x12\xff\x7f\xff\xff")},

	// int32
	{math.MaxInt32, []byte("\x12\xff\xff\xff\x7f")},
	{math.MaxInt32 + 1, []byte("\x13\x00\x00\x00\x80\x00\x00\x00\x00")},
	{math.MinInt32, []byte("\x12\x00\x00\x00\x80")},
	{math.MinInt32 - 1, []byte("\x13\xff\xff\xff\x7f\xff\xff\xff\xff")},

	// int64
	{math.MaxInt64, []byte("\x13\xff\xff\xff\xff\xff\xff\xff\x7f")},
	{math.MinInt64, []byte("\x13\x00\x00\x00\x00\x00\x00\x00\x80")},
}

// Binson BOOLEAN internal representation test data table
var boolTable = []struct {
	val bool
	raw []byte
}{
	{true, []byte("\x44")},
	{false, []byte("\x45")},
}

// Binson DOUBLE internal representation test data table
var doubleTable = []struct {
	val float64
	raw []byte
}{
	{0.0, []byte("\x46\x00\x00\x00\x00\x00\x00\x00\x00")},
	{math.Copysign(0, -1), []byte("\x46\x00\x00\x00\x00\x00\x00\x00\x80")},
	{+3.1415e+10, []byte("\x46\x00\x00\x00\x6f\xeb\x41\x1d\x42")},
	{-3.1415e-10, []byte("\x46\xfc\x17\xac\xd2\x95\x96\xf5\xbd")},
	{math.NaN(), []byte("\x46\x00\x00\x00\x00\x00\x00\xf8\x7f")},
	{math.Inf(+1), []byte("\x46\x00\x00\x00\x00\x00\x00\xf0\x7f")},
	{math.Inf(-1), []byte("\x46\x00\x00\x00\x00\x00\x00\xf0\xff")},
}

func TestTableInts(t *testing.T) {
	for _, record := range intTable {
		var b bytes.Buffer

		// test Encoder
		enc := NewEncoder(&b)
		enc.Integer(record.val)
		enc.Flush()
		if !bytes.Equal(record.raw, b.Bytes()) {
			t.Errorf("Binson int encoder failed: val %v, expected 0x%v != recieved: 0x%v",
				record.val, hex.EncodeToString(record.raw), hex.EncodeToString(b.Bytes()))
		}

		// test Decoder
		var rd = bytes.NewReader(record.raw)
		var dec = NewDecoder(rd)
		typeBeforeValue, _ := rd.ReadByte()
		dec.parseValue(typeBeforeValue, 0)

		if record.val != dec.Value {
			t.Errorf("Binson int decoder failed: expected %v != recieved: %v", record.val, dec.Value)
		}
	}
}

func TestTableBooleans(t *testing.T) {
	for _, record := range boolTable {
		var b bytes.Buffer

		// test Encoder
		enc := NewEncoder(&b)
		enc.Bool(record.val)
		enc.Flush()
		if !bytes.Equal(record.raw, b.Bytes()) {
			t.Errorf("Binson boolean encoder failed: val %v, expected 0x%v != recieved: 0x%v",
				record.val, hex.EncodeToString(record.raw), hex.EncodeToString(b.Bytes()))
		}

		// test Decoder
		var rd = bytes.NewReader(record.raw)
		var dec = NewDecoder(rd)
		typeBeforeValue, _ := rd.ReadByte()
		dec.parseValue(typeBeforeValue, 0)

		if record.val != dec.Value {
			t.Errorf("Binson boolean decoder failed: expected %v != recieved: %v", record.val, dec.Value)
		}
	}
}

func TestTableDoubles(t *testing.T) {
	for _, record := range doubleTable {
		var b bytes.Buffer

		// test Encoder
		enc := NewEncoder(&b)
		enc.Double(record.val)
		enc.Flush()
		if !bytes.Equal(record.raw, b.Bytes()) && !math.IsNaN(record.val) {
			t.Errorf("Binson double encoder failed: val %v, expected 0x%v != recieved: 0x%v",
				record.val, hex.EncodeToString(record.raw), hex.EncodeToString(b.Bytes()))
		}

		// test Decoder
		var rd = bytes.NewReader(record.raw)
		var dec = NewDecoder(rd)
		typeBeforeValue, _ := rd.ReadByte()
		dec.parseValue(typeBeforeValue, 0)

		if record.val != dec.Value && !math.IsNaN(record.val) {
			t.Errorf("Binson double decoder failed: expected %v != recieved: %v", record.val, dec.Value)
		}
	}
}
