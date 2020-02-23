package struc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

type Nested struct {
	Test2 int `struc:"int8"`
}

type Example struct {
	Pad    []byte `struc:"[5]pad"`        // 00 00 00 00 00
	I8f    int    `struc:"int8"`          // 01
	I16f   int    `struc:"int16,big"`         // 00 02
	I32f   int    `struc:"int32,big"`         // 00 00 00 03
	I64f   int    `struc:"int64,big"`         // 00 00 00 00 00 00 00 04
	U8f    int    `struc:"uint8,little"`  // 05
	U16f   int    `struc:"uint16,little"` // 06 00
	U32f   int    `struc:"uint32,little"` // 07 00 00 00
	U64f   int    `struc:"uint64,little"` // 08 00 00 00 00 00 00 00
	Boolf  int    `struc:"bool"`          // 01
	Byte4f []byte `struc:"[4]byte"`       // "abcd"

	I8     int8    // 09
	I16    int16   `struc:"big"`// 00 0a
	I32    int32   `struc:"big"`// 00 00 00 0b
	I64    int64   `struc:"big"`// 00 00 00 00 00 00 00 0c
	U8     uint8   `struc:"little"` // 0d
	U16    uint16  `struc:"little"` // 0e 00
	U32    uint32  `struc:"little"` // 0f 00 00 00
	U64    uint64  `struc:"little"` // 10 00 00 00 00 00 00 00
	BoolT  bool    // 01
	BoolF  bool    // 00
	Byte4  [4]byte // "efgh"
	Float1 float32 `struc:"big"`  // 41 a0 00 00
	Float2 float64 `struc:"big"`  // 41 35 00 00 00 00 00 00

	I32f2 int64 `struc:"int32,big"`  // ff ff ff ff
	U32f2 int64 `struc:"uint32,big"` // ff ff ff ff

	I32f3 int32 `struc:"int64,big"` // ff ff ff ff ff ff ff ff

	Size int    `struc:"sizeof=Str,little"` // 0a 00 00 00
	Str  string `struc:"[]byte"`            // "ijklmnopqr"
	Strb string `struc:"[4]byte"`           // "stuv"

	Size2 int    `struc:"uint8,sizeof=Str2"` // 04
	Str2  string // "1234\0"

	Size3 int    `struc:"uint8,sizeof=Bstr"` // 04
	Bstr  []byte // "5678"

	Size4 int    `struc:"little"`                // 07 00 00 00
	Str4a string `struc:"[]byte,sizefrom=Size4"` // "ijklmno"
	Str4b string `struc:"[]byte,sizefrom=Size4"` // "pqrstuv"

	Size5 int    `struc:"uint8"`          // 04
	Bstr2 []byte `struc:"sizefrom=Size5"` // "5678"

	Nested  Nested  // 00 00 00 01
	NestedP *Nested // 00 00 00 02
	TestP64 *int    `struc:"int64,big"` // 00 00 00 05

	NestedSize int      `struc:"big,sizeof=NestedA"` // 00 00 00 02
	NestedA    []Nested // [00 00 00 03, 00 00 00 04]

	Skip int `struc:"skip"`

	CustomTypeSize    Int3   `struc:"sizeof=CustomTypeSizeArr"` // 00 00 00 04
	CustomTypeSizeArr []byte // "ABCD"
}

var five = 5

var reference = &Example{
	nil,
	1, 2, 3, 4, 5, 6, 7, 8, 0, []byte{'a', 'b', 'c', 'd'},
	9, 10, 11, 12, 13, 14, 15, 16, true, false, [4]byte{'e', 'f', 'g', 'h'},
	20, 21,
	-1,
	4294967295,
	-1,
	10, "ijklmnopqr", "stuv",
	4, "1234",
	4, []byte("5678"),
	7, "ijklmno", "pqrstuv",
	4, []byte("5678"),
	Nested{1}, &Nested{2}, &five,
	6, []Nested{{3}, {4}, {5}, {6}, {7}, {8}},
	0,
	Int3(4), []byte("ABCD"),
}

var referenceBytes = []byte{
	0, 0, 0, 0, 0, // pad(5)
	1, 0, 2, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 4, // fake int8-int64(1-4)
	5, 6, 0, 7, 0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, // fake little-endian uint8-uint64(5-8)
	0,                  // fake bool(0)
	'a', 'b', 'c', 'd', // fake [4]byte

	9, 0, 10, 0, 0, 0, 11, 0, 0, 0, 0, 0, 0, 0, 12, // real int8-int64(9-12)
	13, 14, 0, 15, 0, 0, 0, 16, 0, 0, 0, 0, 0, 0, 0, // real little-endian uint8-uint64(13-16)
	1, 0, // real bool(1), bool(0)
	'e', 'f', 'g', 'h', // real [4]byte
	65, 160, 0, 0, // real float32(20)
	64, 53, 0, 0, 0, 0, 0, 0, // real float64(21)

	255, 255, 255, 255, // fake int32(-1)
	255, 255, 255, 255, // fake uint32(4294967295)

	255, 255, 255, 255, 255, 255, 255, 255, // fake int64(-1)

	10, 0, 0, 0, // little-endian int32(10) sizeof=Str
	'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', // Str
	's', 't', 'u', 'v', // fake string([4]byte)
	04, '1', '2', '3', '4', // real string
	04, '5', '6', '7', '8', // fake []byte(string)

	7, 0, 0, 0, // little-endian int32(7)
	'i', 'j', 'k', 'l', 'm', 'n', 'o', // Str4a sizefrom=Size4
	'p', 'q', 'r', 's', 't', 'u', 'v', // Str4b sizefrom=Size4
	04, '5', '6', '7', '8', // fake []byte(string)

	1, 2, // Nested{1}, Nested{2}
	0, 0, 0, 0, 0, 0, 0, 5, // &five

	0, 0, 0, 6, // int32(6)
	3, 4, 5, 6, 7, 8, // [Nested{3}, ...Nested{8}]

	0, 0, 4, 'A', 'B', 'C', 'D', // Int3(4), []byte("ABCD")
}

func TestCodec(t *testing.T) {
	var buf bytes.Buffer
	if err := Pack(&buf, reference); err != nil {
		t.Fatal(err)
	}
	out := &Example{}
	if err := Unpack(&buf, out); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reference, out) {
		fmt.Printf("got: %#v\nwant: %#v\n", out, reference)
		t.Fatal("encode/decode failed")
	}
}

func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	if err := Pack(&buf, reference); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf.Bytes(), referenceBytes) {
		fmt.Printf("got: %#v\nwant: %#v\n", buf.Bytes(), referenceBytes)
		t.Fatal("encode failed")
	}
}

func TestDecode(t *testing.T) {
	buf := bytes.NewReader(referenceBytes)
	out := &Example{}
	if err := Unpack(buf, out); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reference, out) {
		fmt.Printf("got: %#v\nwant: %#v\n", out, reference)
		t.Fatal("decode failed")
	}
}

func TestSizeof(t *testing.T) {
	size, err := Sizeof(reference)
	if err != nil {
		t.Fatal(err)
	}
	if size != len(referenceBytes) {
		t.Fatalf("sizeof failed; expected %d, got %d", len(referenceBytes), size)
	}
}

type ExampleEndian struct {
	T int `struc:"int16,big"`
}

func TestEndianSwap(t *testing.T) {
	var buf bytes.Buffer
	big := &ExampleEndian{1}
	if err := PackWithOrder(&buf, big, binary.BigEndian); err != nil {
		t.Fatal(err)
	}
	little := &ExampleEndian{}
	if err := UnpackWithOrder(&buf, little, binary.LittleEndian); err != nil {
		t.Fatal(err)
	}
	if little.T != 256 {
		t.Fatal("big -> little conversion failed")
	}
}

func TestNilValue(t *testing.T) {
	var buf bytes.Buffer
	if err := Pack(&buf, nil); err == nil {
		t.Fatal("failed throw error for bad struct value")
	}
	if err := Unpack(&buf, nil); err == nil {
		t.Fatal("failed throw error for bad struct value")
	}
	if _, err := Sizeof(nil); err == nil {
		t.Fatal("failed to throw error for bad struct value")
	}
}

type sliceUnderrun struct {
	Str string   `struc:"[10]byte"`
	Arr []uint16 `struc:"[10]uint16"`
}

func TestSliceUnderrun(t *testing.T) {
	var buf bytes.Buffer
	v := sliceUnderrun{
		Str: "foo",
		Arr: []uint16{1, 2, 3},
	}
	if err := Pack(&buf, &v); err != nil {
		t.Fatal(err)
	}
}


type NestedParent struct {
	Length uint8
	N NestedSliceSize
}


type NestedSliceSize struct {
	*NestedParent `struc:"skip"`
	S []uint8 `struc:"[]uint8,sizefrom=Length"`
}

func NewNestedParent() *NestedParent {
	n := &NestedParent{}
	n.N.NestedParent = n
	return n
}

func TestNestedSlice(t *testing.T) {
	n := NewNestedParent()
	n.N.S = []uint8{
		0x1,
		0x2,
		0x3,
	}
	n.Length = 2
	n.N.NestedParent = n
	var buf bytes.Buffer
	n2 := NewNestedParent()
	if err := Pack(&buf, &n); err != nil {
		t.Fatal(err.Error())
	}
	if err := Unpack(&buf, &n2); err != nil {
		t.Fatal(err.Error())
	}
	if _, err := Sizeof(&n); err != nil {
		t.Fatal(err.Error())
	}
}

type StringSlice struct {
	Length uint8
	S string `struc:"sizefrom=Length"`
}

func TestStringSlice(t *testing.T) {
	v := StringSlice{
		Length: 6,
		S: "Hello, Tester!",
	}
	vBytes := []byte{6, 72, 101, 108, 108, 111, 44}
	var buf bytes.Buffer
	if err := Pack(&buf, &v); err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(vBytes, buf.Bytes()) {
		fmt.Printf("got: %#v\nwant: %#v\n", buf, vBytes)
		t.Fatal("decode failed")
	}
}

func TestStringSlicePadded(t *testing.T) {
	v := StringSlice{
		Length: 20,
		S: "Hello, Tester!",
	}
	vBytes := []byte{20, 72, 101, 108, 108, 111, 44, 32, 84, 101, 115, 116, 101, 114, 33, 0, 0, 0, 0, 0, 0}
	var buf bytes.Buffer
	if err := Pack(&buf, &v); err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(vBytes, buf.Bytes()) {
		fmt.Printf("got: %#v\nwant: %#v\n", buf, vBytes)
		t.Fatal("decode failed")
	}
}

type IntSlice struct {
	Length uint8
	I []uint16 `struc:"sizefrom=Length"`
}

func TestIntSlice(t *testing.T) {
	v := IntSlice{
		Length: 2,
		I: []uint16{0x1122, 0x2233, 0x3344},
	}
	wanted := []byte{2, 0x22, 0x11, 0x33, 0x22}
	var buf bytes.Buffer
	if err := Pack(&buf, &v); err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(wanted, buf.Bytes()) {
		fmt.Printf("got: %#v\nwant: %#v\n", buf, wanted)
		t.Fatal("decode failed")
	}

	var v2 IntSlice
	if err := Unpack(&buf, &v2); err != nil {
		t.Fatal(err.Error())
	}
	var buf2 bytes.Buffer
	if err := Pack(&buf2, &v); err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(wanted, buf2.Bytes()) {
		fmt.Printf("got: %#v\nwant: %#v\n", buf2.Bytes(), wanted)
		t.Fatal("decode failed")
	}
}


func TestIntSlicePadded(t *testing.T) {
	v := IntSlice{
		Length: 4,
		I: []uint16{0x1122, 0x2233},
	}
	wanted := []byte{4, 0x22, 0x11, 0x33, 0x22, 0, 0, 0, 0}
	var buf bytes.Buffer
	if err := Pack(&buf, &v); err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(wanted, buf.Bytes()) {
		fmt.Printf("got: %#v\nwant: %#v\n", buf, wanted)
		t.Fatal("decode failed")
	}

	var v2 IntSlice
	if err := Unpack(&buf, &v2); err != nil {
		t.Fatal(err.Error())
	}
	var buf2 bytes.Buffer
	if err := Pack(&buf2, &v); err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(wanted, buf2.Bytes()) {
		fmt.Printf("got: %#v\nwant: %#v\n", buf2.Bytes(), wanted)
		t.Fatal("decode failed")
	}
}

type PointerSlice struct {
	Length uint8 `struc:"sizeof=I"`
	I []*IntSlice
}

func TestPointerSlice(t *testing.T) {
	v := PointerSlice{
		Length: 2,
		I: []*IntSlice{
			{
				Length: 2,
				I:      []uint16{
					0x00, 0x11,
				},
			},
			{
				Length: 2,
				I:      []uint16{
					0x22, 0x33,
				},
			},
		},
	}
	wanted := []byte{2, 0x2, 0x00, 0x00, 0x11,0x00, 0x2, 0x22,0x00, 0x33, 0x00}
	var buf bytes.Buffer
	if err := Pack(&buf, &v); err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(wanted, buf.Bytes()) {
		fmt.Printf("got: %#v\nwant: %#v\n", buf, wanted)
		t.Fatal("decode failed")
	}

	var v2 PointerSlice
	if err := Unpack(&buf, &v2); err != nil {
		t.Fatal(err.Error())
	}
	var buf2 bytes.Buffer
	if err := Pack(&buf2, &v); err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(wanted, buf2.Bytes()) {
		fmt.Printf("got: %#v\nwant: %#v\n", buf2.Bytes(), wanted)
		t.Fatal("decode failed")
	}
}

type StringSlice2 struct {
	Length    uint32 `struc:"sizeof=Values"`
	Values    []string
	StrLength uint32 `struc:"sizeof=Str"`
	Str       string
	Str2Length uint32 `struc:"sizeof=Str2"`
	Str2      string
}

func TestStringSlice2(t *testing.T) {
	v := StringSlice2{
		Values: []string{
			"Hello",
			"World!",
		},
		StrLength: 2,
		Str: "HW",
		Str2: "HW",
	}
	wanted := []byte{0x2, 0x0, 0x0, 0x0, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x0, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x21, 0x0, 2,0,0,0,0x48, 0x57, 3,0,0,0, 0x48, 0x57, 0x00}
	var buf bytes.Buffer
	if err := Pack(&buf, &v); err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(wanted, buf.Bytes()) {
		fmt.Printf(" got: %#v\nwant: %#v\n", buf.Bytes(), wanted)
		t.Fatal("encode failed")
	}

	var v2 StringSlice2
	if err := Unpack(&buf, &v2); err != nil {
		t.Fatal(err.Error())
	}
	var buf2 bytes.Buffer
	if err := Pack(&buf2, &v); err != nil {
		t.Fatal(err.Error())
	}
	if !reflect.DeepEqual(wanted, buf2.Bytes()) {
		fmt.Printf("got: %#v\nwant: %#v\n", buf2.Bytes(), wanted)
		t.Fatal("decode failed")
	}
}