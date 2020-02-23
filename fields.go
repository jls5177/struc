package struc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Fields []*Field

func (f Fields) SetByteOrder(order binary.ByteOrder) {
	for _, field := range f {
		if field != nil {
			field.Order = order
		}
	}
}

func (f Fields) String() string {
	fields := make([]string, len(f))
	for i, field := range f {
		if field != nil {
			fields[i] = field.String()
		}
	}
	return "{" + strings.Join(fields, ", ") + "}"
}

func (f Fields) Sizeof(val reflect.Value, options *Options) int {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	size := 0
	for i, field := range f {
		if field != nil {
			var sliceLength int
			// Grab the size in the from field if one was specified
			if field.Sizefrom != nil {
				if n, ok := SizeFromField(val.FieldByIndex(field.Sizefrom)); ok {
					sliceLength = n
				}
			}
			size += field.Size(val.Field(i), options, sliceLength)
		}
	}
	return size
}

func SizeFromField(field reflect.Value) (int, bool) {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(field.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := int(field.Uint())
		// all the builtin array length types are native int
		// so this guards against weird truncation
		if n < 0 {
			return 0, true
		}
		return n, true
	default:
		return 0, false
	}
}

func (f Fields) sizefrom(val reflect.Value, index []int) int {
	field := val.FieldByIndex(index)
	if n, ok := SizeFromField(field); ok {
		return n
	}
	name := val.Type().FieldByIndex(index).Name
	panic(fmt.Sprintf("sizeof field %T.%s not an integer type", val.Interface(), name))
}

func (f Fields) Pack(buf []byte, val reflect.Value, options *Options) (int, error) {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	pos := 0
	for i, field := range f {
		if field == nil {
			continue
		}
		v := val.Field(i)
		length := field.Len
		if field.Sizefrom != nil {
			length = f.sizefrom(val, field.Sizefrom)
		}
		if length <= 0 && field.Slice {
			length = v.Len()
		}
		if field.Sizeof != nil {
			length := val.FieldByIndex(field.Sizeof).Len()
			sizeofField := f[field.Sizeof[0]]
			if sizeofField.NullString {
				length += 1
			}
			switch field.kind {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				// allocating a new int here has fewer side effects (doesn't update the original struct)
				// but it's a wasteful allocation
				// the old method might work if we just cast the temporary int/uint to the target type
				v = reflect.New(v.Type()).Elem()
				v.SetInt(int64(length))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				v = reflect.New(v.Type()).Elem()
				v.SetUint(uint64(length))
			default:
				panic(fmt.Sprintf("sizeof field is not int or uint type: %s, %s", field.Name, v.Type()))
			}
		}
		if n, err := field.Pack(buf[pos:], v, length, options); err != nil {
			return n, err
		} else {
			pos += n
		}
	}
	return pos, nil
}

func (f Fields) Unpack(r io.Reader, val reflect.Value, options *Options) error {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	var tmp [8]byte
	var buf []byte
	for i, field := range f {
		if field == nil {
			continue
		}
		v := val.Field(i)
		length := field.Len
		if field.Sizefrom != nil {
			length = f.sizefrom(val, field.Sizefrom)
		}
		if v.Kind() == reflect.Ptr && !v.Elem().IsValid() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if field.Type == Struct {
			if field.Slice {
				vals := reflect.MakeSlice(v.Type(), length, length)
				for i := 0; i < length; i++ {
					v := vals.Index(i)

					// create a new element to unpack into if have a pointer slice
					if v.Kind() == reflect.Ptr && !v.Elem().IsValid() {
						v.Set(reflect.New(v.Type().Elem()))
					}

					fields, err := parseFields(v)
					if err != nil {
						return err
					}
					if err := fields.Unpack(r, v, options); err != nil {
						return err
					}
				}
				v.Set(vals)
			} else {
				// TODO: DRY (we repeat the inner loop above)
				fields, err := parseFields(v)
				if err != nil {
					return err
				}
				if err := fields.Unpack(r, v, options); err != nil {
					return err
				}
			}
			continue
		} else {
			typ := field.Type.Resolve(options)
			if typ == CustomType {
				if err := v.Addr().Interface().(Custom).Unpack(r, length, options); err != nil {
					return err
				}
			} else if typ == String {
				if field.Slice {
					vals := reflect.MakeSlice(v.Type(), length, length)
					for i := 0; i < length; i++ {
						v := vals.Index(i)

						// create a new element to unpack into if have a pointer slice
						if v.Kind() == reflect.Ptr && !v.Elem().IsValid() {
							v.Set(reflect.New(v.Type().Elem()))
						}

						s := readString(r, -1)
						v.SetString(s)
					}
					v.Set(vals)
				} else {
					max := -1
					if field.Sizefrom != nil && length != 0 {
						max = length
					}
					s := readString(r, max)
					v.SetString(s)
				}
			} else {
				size := length * field.Type.Resolve(options).Size()
				if size < 8 {
					buf = tmp[:size]
				} else {
					buf = make([]byte, size)
				}
				if _, err := io.ReadFull(r, buf); err != nil {
					return err
				}
				err := field.Unpack(buf[:size], v, length, options)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// readString reads a string byte by byte until either max characters are
// read or we reach a null string (if max == -1).
func readString(r io.Reader, max int) string {
	stringBuf := bytes.Buffer{}
	if max == 0 {
		return ""
	}

	var err error
	var n int
	b := make([]uint8, 1)
	for {
		n, err = r.Read(b)
		if err == io.EOF || n != 1 {
			break
		} else if max < 0 && b[0] == 0 {
			break
		}
		stringBuf.Write(b)
		if max > 0 && stringBuf.Len() >= max {
			break
		}
	}
	return string(stringBuf.Bytes())
}