package struc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const (
	Apples  = 1 << 0
	Oranges = 1 << 1
	Grapes  = 1 << 4
)

type SupportedFruits struct {
	Bitmap
}

func (b *SupportedFruits) GetMap() BitmapperType {
	return ConvertEnum(map[string]uint64{
		"APPLES":  Apples,
		"ORANGES": Oranges,
		"GRAPES":  Grapes,
	})
}

type FruitOptions struct {
	Bitmap
}

const (
	SkinNotPeeled = 2
	SkinPeeled = 3
)
func (b *FruitOptions) GetMap() BitmapperType {
	return ConvertEnum(map[string]uint64{
		"Skin Not Peeled": SkinNotPeeled,
		"Skin Peeled":     SkinPeeled,
	})
}

func (b *FruitOptions) Value() (uint64, error) {
	return b.Bitmap.Value(b)
}

type SupportedFruitsTable struct {
	SupportedFruits SupportedFruits `struc:"uint32" json:",omitempty"`
	FruitOptions    *FruitOptions   `struc:"uint32" json:",omitempty"`
}

func TestFruitUnderstanding(t *testing.T) {
	var sample SupportedFruitsTable

	var data = `SupportedFruits:
- APPLES
- ORANGES
- GRAPES
FruitOptions: Skin Peeled`

	// Convert YAML into structure
	if err := yaml.Unmarshal([]byte(data), &sample); err != nil {
		t.Errorf("fail!fail!fail!: %v", err.Error())
	}

	// Pack the structure into bytes
	var buf bytes.Buffer
	if err := Pack(&buf, &sample); err != nil {
		t.Errorf("fail!fail!fail!: %v", err.Error())
	}
	fmt.Printf("Value: %X\n", buf.Bytes())

	// Verify Unpack works by converting bytes into structure
	unsample := SupportedFruitsTable{
		SupportedFruits: SupportedFruits{},
		FruitOptions:    &FruitOptions{},
	}
	if err := Unpack(&buf, &unsample); err != nil {
		t.Errorf("fail!fail!fail!: %v", err.Error())
	}

	// Sort the []strings and perform a deep equality check
	sort.Strings(sample.SupportedFruits.Values)
	sort.Strings(unsample.SupportedFruits.Values)
	if !reflect.DeepEqual(sample.SupportedFruits, unsample.SupportedFruits) {
		t.Errorf("fail!fail!fail!: Pack and Unpack not equal")
	}
}

func TestFruitUnderstanding2(t *testing.T) {
	var sample SupportedFruitsTable

	var data = `SupportedFruits:
- APPLES
- ORANGES
- GRAPES`

	// Convert YAML into structure
	if err := yaml.Unmarshal([]byte(data), &sample); err != nil {
		t.Errorf("fail!fail!fail!: %v", err.Error())
	}

	// Pack the structure into bytes
	var buf bytes.Buffer
	if err := Pack(&buf, &sample); err != nil {
		t.Errorf("fail!fail!fail!: %v", err.Error())
	}
	fmt.Printf("Value: %X\n", buf.Bytes())

	// Verify Unpack works by converting bytes into structure
	unsample := SupportedFruitsTable{
		SupportedFruits: SupportedFruits{},
		FruitOptions:    &FruitOptions{},
	}
	if err := Unpack(&buf, &unsample); err != nil {
		t.Errorf("fail!fail!fail!: %v", err.Error())
	}

	// Sort the []strings and perform a deep equality check
	sort.Strings(sample.SupportedFruits.Values)
	sort.Strings(unsample.SupportedFruits.Values)
	if !reflect.DeepEqual(sample.SupportedFruits, unsample.SupportedFruits) {
		t.Errorf("fail!fail!fail!: Pack and Unpack not equal")
	}
}

func TestFruitMarshaling(t *testing.T) {
	var sample SupportedFruitsTable

	var data = `SupportedFruits:
- GRAPES
- ORANGES`

	// Convert YAML into structure
	if err := yaml.Unmarshal([]byte(data), &sample); err != nil {
		t.Errorf("fail!fail!fail!: %v", err.Error())
	}

	if v, err := json.Marshal(&sample); err == nil {
		fmt.Printf(string(v))
	}

	// Now go backawards back to YAML and compare
	sample2 := SupportedFruitsTable{
		SupportedFruits: SupportedFruits{
			Bitmap{
				Values: []string{
					"GRAPES",
					"ORANGES",
				},
			},
		},
	}
	var buf2 []byte
	buf2, err := yaml.Marshal(&sample2)
	if err != nil {
		t.Errorf("fail!fail!fail!: %v", err.Error())
	} else {
		fmt.Printf("Return vaule is: \n%v", string(buf2))
	}

	data2 := string(buf2)
	if strings.TrimSpace(data) != strings.TrimSpace(data2) {
		t.Errorf("fail!fail!fail!: Unmarshal != Marshal are not the same")
	}
}

func TestFruitStringBitmap(t *testing.T) {
	var sample SupportedFruitsTable

	var data = `SupportedFruits: ORANGES`

	// Convert YAML into structure
	if err := yaml.Unmarshal([]byte(data), &sample); err != nil {
		t.Errorf("fail!fail!fail!: %v", err.Error())
	}

	if v, err := json.Marshal(&sample); err == nil {
		fmt.Printf(string(v))
	}

	// Now go backawards back to YAML and compare
	sample2 := SupportedFruitsTable{
		SupportedFruits: SupportedFruits{
			Bitmap{
				Values: []string{
					"ORANGES",
				},
			},
		},
	}
	var buf2 []byte
	buf2, err := yaml.Marshal(&sample2)
	if err != nil {
		t.Errorf("fail!fail!fail!: %v", err.Error())
	} else {
		fmt.Printf("Return vaule is: \n%v", string(buf2))
	}

	data2 := string(buf2)
	if strings.TrimSpace(data) != strings.TrimSpace(data2) {
		t.Errorf("fail!fail!fail!: Unmarshal != Marshal are not the same")
	}
}

func TestBitmapToValue(t *testing.T) {
	bitmap := SupportedFruits{
		Bitmap{
			Values: []string{
				"ORANGES",
			},
		},
	}

	value, err := bitmap.Value(&bitmap)
	if err != nil {
		t.Error(err)
	}
	if value != Oranges {
		t.Errorf("invalid value: %v != %v", value, Oranges)
	}

	value, err = BitmapValue(&bitmap)
	if err != nil {
		t.Error(err)
	}
	if value != Oranges {
		t.Errorf("invalid value: %v != %v", value, Oranges)
	}
}

type BitmapRange struct {
	Bitmap
}

func (b *BitmapRange) GetMap() BitmapperType {
	return ConvertEnum(map[string]uint64{
		"APPLES":  Apples,
		"ORANGES": Oranges,
		"GRAPES":  Grapes,
	}).AppendRange(0,8, 1)
}

func TestBitmapRange(t *testing.T) {
	bitmap := BitmapRange{
		Bitmap{
			Values: []string{
				"5",
			},
		},
	}

	value, err := bitmap.Value(&bitmap)
	if err != nil {
		t.Error(err)
	}
	if value != 5 {
		t.Errorf("invalid value: %v != %v", value, Oranges)
	}
}
