// Package binson implements s small, high-performance implementation of Binson, see binson.org.
package binson

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// ValueType is type signature for each binson item
type ValueType uint

// Binson item types enumeration
const (
	Boolean ValueType = iota
	Integer
	Double
	String
	Bytes
	Array
	Object
)

// Binson item signatures
const (
	sigBegin      byte = 0x40
	sigEnd        byte = 0x41
	sigBeginArray byte = 0x42
	sigEndArray   byte = 0x43
	sigTrue       byte = 0x44
	sigFalse      byte = 0x45
	sigInteger1   byte = 0x10
	sigInteger2   byte = 0x11
	sigInteger4   byte = 0x12
	sigInteger8   byte = 0x13
	sigDouble     byte = 0x46
	sigString1    byte = 0x14
	sigString2    byte = 0x15
	sigString4    byte = 0x16
	sigBytes1     byte = 0x18
	sigBytes2     byte = 0x19
	sigBytes4     byte = 0x1a
)

const intLengthMask byte = 0x03

const oneByte byte = 0x00
const twoBytes byte = 0x01
const fourBytes byte = 0x02
const eightBytes byte = 0x03

//var intSizeMap = [...]int{oneByte: 1, twoBytes: 2, fourBytes: 4, eightBytes: 8}

const twoTo7 int64 = 128
const twoTo15 int64 = 32768
const twoTo31 int64 = 2147483648

// Binson Decoder private constants
const (
	stateZero = iota
	stateBeforeField
	stateBeforeArrayValue
	stateBeforeArray
	stateEndOfArray
	stateBeforeObject
	stateEndOfObject
)

// A Decoder represents an Binson parser reading a particular input stream.
type Decoder struct {
	r   *bufio.Reader
	err error

	Name      string
	Value     interface{}
	ValueType ValueType

	state   int
	sigByte byte // temp
}

// NewDecoder creates a new binson parser reading from r.
// If r does not implement io.ByteReader, NewDecoder will
// do its own buffering.
func NewDecoder(r io.Reader) *Decoder {
	d := &Decoder{r: bufio.NewReader(r), state: stateZero}
	return d
}

// Field parses until an expected field with the given name is found
// (without considering fields of inner objects).
func (d *Decoder) Field(name string) bool {
	for d.NextField() {
		if d.err != nil {
			return false
		}
		if name == d.Name {
			return true
		}
	}

	return false
}

// NextField reads next field, returns true if a field was found and false
// if end-of-object was reached.
// If  boolean/integer/double/bytes/string was found, the value is also read
// and is available in `Value` field
func (d *Decoder) NextField() bool {
	switch d.state {
	case stateZero:
		d.parseBegin()
	case stateEndOfObject:
		d.err = fmt.Errorf("reached end-of-object")
		return false
	case stateBeforeObject:
		d.state = stateBeforeField
		for d.NextField() {
			if d.err != nil {
				return false
			}
		}
		d.state = stateBeforeField
	case stateBeforeArray:
		d.state = stateBeforeArrayValue
		for d.NextArrayValue() {
			if d.err != nil {
				return false
			}
		}
		d.state = stateBeforeField
	}

	if d.state != stateBeforeField {
		d.err = fmt.Errorf("not ready to read a field, state: %v", d.state)
		return false
	}

	typeBeforeName, err := d.r.ReadByte()
	if err != nil {
		d.err = fmt.Errorf("abnormal end of input stream detected")
		return false
	}
	if typeBeforeName == sigEnd {
		d.state = stateEndOfObject
		return false
	}
	d.parseFieldName(typeBeforeName)

	typeBeforeValue, err := d.r.ReadByte()
	if err != nil {
		d.err = fmt.Errorf("abnormal end of input stream detected")
		return false
	}
	d.parseValue(typeBeforeValue, stateBeforeField)

	return true
}

// NextArrayValue reads next binson ARRAY value,
// returns true if a field was found and false, if end-of-object was reached.
// If boolean/integer/double/bytes/string was found, the value is also read
// and is available in `Value` field
func (d *Decoder) NextArrayValue() bool {
	if d.state == stateBeforeArray {
		d.state = stateBeforeArrayValue
		for d.NextArrayValue() {
			if d.err != nil {
				return false
			}
		}
		d.state = stateBeforeArrayValue
	}

	if d.state == stateBeforeObject {
		d.state = stateBeforeField
		for d.NextField() {
			if d.err != nil {
				return false
			}
		}
		d.state = stateBeforeArrayValue
	}

	if d.state != stateBeforeArrayValue {
		d.err = fmt.Errorf("not before array value: %v", d.state)
		return false
	}

	sig, err := d.r.ReadByte()
	if err != nil {
		d.err = fmt.Errorf("abnormal end of input stream detected")
		return false
	}
	if sig == sigEndArray {
		d.state = stateEndOfArray
		return false
	}
	d.parseValue(sig, stateBeforeArrayValue)

	return true
}

// GoIntoObject navigates decoder inside the expected OBJECT
func (d *Decoder) GoIntoObject() {
	if d.state != stateBeforeObject {
		d.err = fmt.Errorf("unexpected parser state, not an object field")
		return
	}
	d.state = stateBeforeField
}

// GoIntoArray navigates decoder inside the expected ARRAY
func (d *Decoder) GoIntoArray() {
	if d.state != stateBeforeArray {
		d.err = fmt.Errorf("unexpected parser state, not an array field")
		return
	}
	d.state = stateBeforeArrayValue
}

// GoUpToObject navigates decoder to the parent OBJECT
func (d *Decoder) GoUpToObject() {
	if d.state == stateBeforeArrayValue {
		for d.NextArrayValue() {
			if d.err != nil {
				return
			}
		}
	}

	if d.state == stateBeforeField {
		for d.NextField() {
			if d.err != nil {
				return
			}
		}
	}

	if d.state != stateEndOfObject && d.state != stateEndOfArray {
		d.err = fmt.Errorf("unexpected parser state: %v", d.state)
		return
	}

	d.state = stateBeforeField
}

// GoUpToArray navigates decoder to the parent ARRAY
func (d *Decoder) GoUpToArray() {
	if d.state == stateBeforeArrayValue {
		for d.NextArrayValue() {
			if d.err != nil {
				return
			}
		}
	}

	if d.state == stateBeforeField {
		for d.NextField() {
			if d.err != nil {
				return
			}
		}
	}

	if d.state != stateEndOfObject && d.state != stateEndOfArray {
		d.err = fmt.Errorf("unexpected parser state: %v", d.state)
		return
	}

	d.state = stateBeforeArrayValue
}

/* === private methods === */

func (d *Decoder) parseValue(sigByte byte, afterValueState int) {
	switch sigByte {
	case sigBegin:
		d.ValueType = Object
		d.state = stateBeforeObject
	case sigBeginArray:
		d.ValueType = Array
		d.state = stateBeforeArray
	case sigFalse, sigTrue:
		d.ValueType = Boolean
		d.Value = sigByte == sigTrue
		d.state = afterValueState
	case sigDouble:
		var d64 float64
		d.ValueType = Double
		d.err = binary.Read(d.r, binary.LittleEndian, &d64)
		d.Value = d64
		d.state = afterValueState
	case sigInteger1, sigInteger2, sigInteger4, sigInteger8:
		d.ValueType = Integer
		d.Value = d.parseInteger(sigByte)
		d.state = afterValueState
	case sigString1, sigString2, sigString4:
		d.ValueType = String
		d.Value = d.parseStringBytes(sigByte)
		d.state = afterValueState
	case sigBytes1, sigBytes2, sigBytes4:
		d.ValueType = Bytes
		d.Value = d.parseStringBytes(sigByte)
		d.state = afterValueState
	default:
		d.err = fmt.Errorf("Unexpected type byte: %v", sigByte)
	}
}

func (d *Decoder) parseFieldName(sigBeforeName byte) {
	switch sigBeforeName {
	case sigString1, sigString2, sigString4:
		d.Name = d.parseStringBytes(sigBeforeName).(string)
	default:
		d.err = fmt.Errorf("unexpected type: %v", sigBeforeName)
	}
}

func (d *Decoder) parseBegin() {
	d.err = binary.Read(d.r, binary.LittleEndian, &d.sigByte)
	if d.sigByte != sigBegin {
		d.err = fmt.Errorf("Expected BEGIN, got: %v", d.sigByte)
		return
	}
	d.state = stateBeforeField
}

func (d *Decoder) parseStringBytes(sigByte byte) interface{} {
	ln := d.parseInteger(sigByte)

	if ln < 0 {
		d.err = fmt.Errorf("Bad string/bytes length: %v", ln)
		return nil
	}

	if ln > 10*1000000 {
		d.err = fmt.Errorf("String/Bytes length too big: %v", ln)
		return nil
	}

	var buf = make([]byte, ln)
	n, err := io.ReadFull(d.r, buf)

	if int64(n) < ln || err != nil {
		d.err = fmt.Errorf("abnormal end of input stream detected")
		return nil
	}

	if sigByte >= sigBytes1 {
		return buf
	}
	return string(buf)
}

func (d *Decoder) parseInteger(sigByte byte) int64 {
	switch sigByte & intLengthMask {
	case oneByte:
		var i1 int8
		d.err = binary.Read(d.r, binary.LittleEndian, &i1)
		return int64(i1)
	case twoBytes:
		var i2 int16
		d.err = binary.Read(d.r, binary.LittleEndian, &i2)
		return int64(i2)
	case fourBytes:
		var i4 int32
		d.err = binary.Read(d.r, binary.LittleEndian, &i4)
		return int64(i4)
	case eightBytes:
		var i8 int64
		d.err = binary.Read(d.r, binary.LittleEndian, &i8)
		return int64(i8)
	default:
		d.err = fmt.Errorf("never happens")
		return -1
	}
}

// An Encoder writes binson data to an output stream.
type Encoder struct {
	w   *bufio.Writer
	err error
}

// NewEncoder returns a new encoder that writes to w, with buffering
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: bufio.NewWriter(w)}
}

// Flush encoder buffers
func (e *Encoder) Flush() {
	e.w.Flush()
}

// Begin writes OBJECT begin signature to output stream
func (e *Encoder) Begin() {
	e.err = e.w.WriteByte(sigBegin)
}

// End writes OBJECT end signature to output stream
func (e *Encoder) End() {
	e.err = e.w.WriteByte(sigEnd)
}

// BeginArray writes ARRAY begin signature to output stream
func (e *Encoder) BeginArray() {
	e.err = e.w.WriteByte(sigBeginArray)
}

// EndArray writes ARRAY end signature to output stream
func (e *Encoder) EndArray() {
	e.err = e.w.WriteByte(sigEndArray)
}

// Bool writes specified boolean value to output stream
func (e *Encoder) Bool(val bool) {
	var sig = sigTrue
	if !val {
		sig = sigFalse
	}
	e.err = e.w.WriteByte(sig)
}

// Integer writes specified integer value to output stream
func (e *Encoder) Integer(val int64) {
	e.writeIntegerOrLength(sigInteger1, val)
}

// Double writes float64 value to output stream
func (e *Encoder) Double(val float64) {
	e.w.WriteByte(sigDouble)
	binary.Write(e.w, binary.LittleEndian, val)
}

// String writes string value to output stream
func (e *Encoder) String(val string) {
	e.writeIntegerOrLength(sigString1, int64(len(val)))
	e.w.WriteString(val)
}

// Bytes writes []byte value to output stream
func (e *Encoder) Bytes(val []byte) {
	e.writeIntegerOrLength(sigBytes1, int64(len(val)))
	e.w.Write(val)
}

// Name writes string value as OBJECT item's name to output stream
func (e *Encoder) Name(val string) {
	e.String(val)
}

/* === private methods === */

func (e *Encoder) writeIntegerOrLength(baseType byte, val int64) {
	switch {
	case val >= -twoTo7 && val < twoTo7:
		e.w.WriteByte(baseType | oneByte)
		binary.Write(e.w, binary.LittleEndian, byte(val))
	case val >= -twoTo15 && val < twoTo15:
		e.w.WriteByte(baseType | twoBytes)
		binary.Write(e.w, binary.LittleEndian, int16(val))
	case val >= -twoTo31 && val < twoTo31:
		e.w.WriteByte(baseType | fourBytes)
		binary.Write(e.w, binary.LittleEndian, int32(val))
	default:
		e.w.WriteByte(baseType | eightBytes)
		binary.Write(e.w, binary.LittleEndian, int64(val))
	}
}
