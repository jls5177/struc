package struc

import (
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type SupportedFruits struct {
	Bitmap
}
func (b *SupportedFruits) GetMap() BitmapperType {
	return ConvertBitmap(map[string]uint64{
		"APPLES":  0,
		"ORANGES": 1,
		"GRAPES":  4,
	})
}

type FruitOptions struct {
	Bitmap
}
func (b *FruitOptions) GetMap() BitmapperType {
	return ConvertEnum(map[string]uint64{
		"Skin Not Peeled": 2,
		"Skin Peeled":     3,
	})
}

type SupportedFruitsTable struct {
	SupportedFruits SupportedFruits `struc:"uint32" json:",omitempty"`
	FruitOptions *FruitOptions `struc:"uint32" json:",omitempty"`
}

func TestFruitUnderstanding(t *testing.T) {
	var sample SupportedFruitsTable

	var data = `SupportedFruits:
- APPLES
- ORANGES
- GRAPES
FruitOptions:
 - Skin Peeled`

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
