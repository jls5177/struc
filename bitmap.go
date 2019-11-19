package struc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
)

type BitmapperType = map[string]uint64

type Bitmapper interface {
	GetMap() BitmapperType
}

// ConvertBitmap converts a bitmap into a common BitmapperType with bit values converted
func ConvertBitmap(b map[string]uint64) BitmapperType {
	return convertToBitmapper(true, b)
}

// ConvertBitmap converts a enum map into a common BitmapperType with values not converted
func ConvertEnum(b map[string]uint64) BitmapperType {
	return convertToBitmapper(false, b)
}

// convertToBitmapper is an internal function that handles either bitmaps or enum values
func convertToBitmapper(isBitmap bool, b map[string]uint64) BitmapperType {
	convMap := BitmapperType{}
	for k, v := range b {
		if isBitmap {
			convMap[k] = 1 << v
		} else {
			convMap[k] = v
		}
	}
	return convMap
}

// Bitmap is a base type for representing Bitmap values
type Bitmap struct {
	Values []string
}

// MarshalJSON allows for marshaling a Bitmap into JSON
func (b *Bitmap) MarshalJSON() ([]byte, error) {
	if b.Values == nil {
		return []byte(""), nil
	}
	if len(b.Values) == 1 {
		value := b.Values[0]
		return json.Marshal(&value)
	}
	return json.Marshal(&b.Values)
}

// BitmapValue returns the integer value for the given structure that embeds the Bitmap type
func BitmapValue(bitmapper Bitmapper) (uint64, error) {
	// use reflection to get the bitmap structure member
	val := reflect.ValueOf(bitmapper)
	return bitmapValue(val, bitmapper)
}

func bitmapValue(val reflect.Value, bitmapper Bitmapper) (uint64, error) {
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0, nil
		}
		val = reflect.Indirect(val)
	}
	field := val.FieldByName("Bitmap")
	if !field.IsValid() {
		return 0, fmt.Errorf("type is missing embedded Bitmap structure: %+v\n", val.Type())
	}

	bitmap, ok := field.Interface().(Bitmap)
	if !ok {
		return 0, fmt.Errorf("missing embedded Bitmap structure: %+v", val.Type())
	}
	return bitmap.Value(bitmapper)
}

// Value converts the value corresponding to the the flags set. It is not possible to
// get the parent structure from an embedded type using reflection. So make the user
// pass the structure in.
func (b *Bitmap) Value(bitmapper Bitmapper) (uint64, error) {
	var err error

	// Build the bitmap value, but first create a case-insensitive bitmap map
	bitmap := make(BitmapperType, len(bitmapper.GetMap()))
	for k, v := range bitmapper.GetMap() {
		bitmap[strings.ToLower(k)] = v
	}

	var n uint64
	for _, v := range b.Values {
		if b, valid := bitmap[strings.ToLower(v)]; valid {
			n |= b
		} else if v != "" {
			err = errors.New(fmt.Sprintf("invalid bitmap value: %s", v))
			break
		}
	}
	return n, err
}

func removeDuplicatesFromSlice(s []string) []string {
	m := make(map[string]bool)
	for _, item := range s {
		if _, ok := m[item]; !ok {
			m[item] = true
		}
	}

	var result []string
	for item, _ := range m {
		result = append(result, item)
	}
	return result
}

// UnmarshalJSON allows for converting JSON into a Bitmap
func (b *Bitmap) UnmarshalJSON(data []byte) error {
	var values []string

	// Determine if the value is a list of strings to properly decode the value
	var re = regexp.MustCompile(`(?m)\[.*?\]`)
	if re.Match(data) {
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}

		b.Values = append(b.Values, values...)
	} else {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		b.Values = append(b.Values, value)
	}

	// Remove all duplicates from the slice to prevent issues when packing the value
	b.Values = removeDuplicatesFromSlice(b.Values)
	return nil
}

// bitmapPack converts a Bitmap structure into a slice of bytes with the correct endian order based on
// the user selected type in the field tag
func bitmapPack(buf []byte, val reflect.Value, length int, options *Options, f *Field) (int, error) {
	if f.Bitmap == nil {
		return 0, fmt.Errorf("invalid type, not a Bitmap structure")
	}

	typ := f.Type.Resolve(options)
	size := typ.Size()
	byteCount := length * size

	// Extract the set values from the given Bitmap structure
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			// If the pointer is nil then return and leave the bytes equal to zero
			return byteCount, nil
		}
		val = reflect.Indirect(val)
	}
	bitmapper := val.Addr().Interface().(Bitmapper)

	n, err := bitmapValue(val, bitmapper)
	if err != nil {
		log.Panic(err.Error())
	}

	// Convert the uint64 into the requested size
	order := f.Order
	if options.Order != nil {
		order = options.Order
	}
	for i, pos := 0, 0; i < length; i, pos = i+1, size {
		switch typ {
		case Bool:
			if n != 0 {
				buf[pos] = 1
			} else {
				buf[pos] = 0
			}
		case Int8, Uint8:
			buf[pos] = byte(n)
		case Int16, Uint16:
			order.PutUint16(buf[pos:], uint16(n))
		case Int32, Uint32:
			order.PutUint32(buf[pos:], uint32(n))
		case Int64, Uint64:
			order.PutUint64(buf[pos:], uint64(n))
		}
		n = n >> (8 * uint(size))
	}

	return byteCount, nil
}

func bitmapUnpack(buf []byte, val reflect.Value, length int, options *Options, f *Field) error {
	if f.Bitmap == nil {
		return fmt.Errorf("invalid type, not a Bitmap structure")
	}

	typ := f.Type.Resolve(options)
	size := typ.Size()
	//byteCount := length * size

	order := f.Order
	if options.Order != nil {
		order = options.Order
	}

	var bitmapValue uint64
	for i, pos := 0, 0; i < length; i, pos = i+1, size {
		var n uint64
		switch typ {
		case Bool:
			if buf[pos] != 0 {
				n = 1
			} else {
				n = 0
			}
		case Int8, Uint8:
			n = uint64(buf[pos])
		case Int16, Uint16:
			n = uint64(order.Uint16(buf[pos:]))
		case Int32, Uint32:
			n = uint64(order.Uint32(buf[pos:]))
		case Int64, Uint64:
			n = order.Uint64(buf[pos:])
		}
		bitmapValue |= n << (8 * uint(pos))
	}

	// Now that we have the value lets find all set flags

	// If there are no values set then nothing else to do
	if bitmapValue == 0 {
		return nil
	}

	// Build list of enumeration values that were set
	var setValues []string
	for bitmask, value := range f.Bitmap {
		if (value & bitmapValue) == value {
			setValues = append(setValues, bitmask)
		}
	}

	// Warning: Crazy magic ahead, not for the faint of heart
	// using reflection to get the Struct->Bitmap->Values field from the passed in value
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			// If the pointer is nil then return and leave the bytes equal to zero
			return nil
		}
		val = reflect.Indirect(val)
	}
	// get the Bitmap field to make sure we get the right Values field
	bitmapField := val.FieldByName("Bitmap")
	if !bitmapField.IsValid() {
		return fmt.Errorf("type is missing embedded Bitmap structure: %+v\n", val.Type())
	}
	// now get the Values field and make sure it is a slice
	valuesField := bitmapField.FieldByName("Values")
	if valuesField.IsValid() {
		if valuesField.Kind() != reflect.Slice {
			return fmt.Errorf("reflection error: did not see a slice looking back at me")
		}
		// The Values slice needs to be resized to match the new values
		valuesField.Set(reflect.MakeSlice(reflect.TypeOf(setValues), len(setValues), len(setValues)))
		// now copy over all of the values built above
		for i, w := range setValues {
			valuesField.Index(i).Set(reflect.ValueOf(w))
		}
	}

	return nil
}
