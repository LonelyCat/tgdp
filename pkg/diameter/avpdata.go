//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: avpdata.go
// Description: Diameter pkg: AVP global data processing
//

package diameter

import (
	"encoding/hex"
	"iter"
	"log/slog"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// -- Consts
// --
const (
	AVP_DATA_APPEND  = -1
	AVP_DATA_REPLACE = -2
	AVP_DATA_CLEANUP = -3
)

// -- Types
// --
type AvpDataStore map[uint32][]*AvpData

// TODO:
// type AvpSpecFunc func(data any) []*AvpData

// -- Variables
// --
var (
	globalAvpDataStore AvpDataStore
	// encSpecFuncs map[int]AvpSpecFunc
	// decSpecFuncs map[int]AvpSpecFunc
)

// -- Init
// --
func init() {
	globalAvpDataStore = make(AvpDataStore)
}

// -- Functions
// --
func FetchAvpValue(avpId any, store *AvpDataStore) ([]*AvpData, error) {
	avp, err := Dict.GetAvp(avpId)
	if err != nil {
		return nil, err
	}

	if store == nil {
		store = &globalAvpDataStore
	}

	if data, exists := (*store)[avp.Code]; exists {
		return data, nil
	}

	return nil, nil
}

func FetchAvpValue2(avpId any, store *AvpDataStore) (*Avp, []*AvpData, error) {
	avp, err := Dict.GetAvp(avpId)
	if err != nil {
		return nil, nil, err
	}

	if store == nil {
		store = &globalAvpDataStore
	}

	if data, exists := (*store)[avp.Code]; exists {
		return avp, data, nil
	}

	return avp, nil, nil
}

func FetchAvpsValues(rules []*AvpRule, avps *[]*Avp, store *AvpDataStore) (uint32, error) {
	if store == nil {
		store = &globalAvpDataStore
	}

	totalSize := uint32(0)

	for _, rule := range rules {
		avp, err := Dict.GetAvpByName(rule.Name)
		if err != nil {
			if rule.Required {
				return 0, err
			} else {
				continue
			}
		}

		data, err := FetchAvpValue(avp.Code, store)
		if err != nil {
			return 0, err
		}

		if data != nil {
			for _, d := range data {
				cloned := avp.clone()

				if avp.IsGrouped() {
					members := make([]*Avp, 0)
					s := d.Value.(*AvpDataStore)
					size, err := FetchAvpsValues(avp.Group.Members, &members, s)
					if err != nil {
						return 0, err
					}
					cloned.Data = new(AvpData)
					cloned.Data.Value = members
					cloned.Data.Size = size
					cloned.Length = alignTo4(cloned.Len())
					*avps = append(*avps, cloned)
					totalSize += cloned.Length
					continue
				}

				cloned.Data = d
				cloned.Len()
				*avps = append(*avps, cloned)
				totalSize += cloned.Len()
			}
		} else if rule.Required {
			err := ErrNoValueForReqAvp{avp.Name}
			slog.Error(err.Error())
		}
	}

	return totalSize, nil
}

func LoadAvpsDataFromYaml(yamlFile string, index int) error {
	yamlText, err := os.ReadFile(yamlFile)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return AvpDataFromYamlStr(string(yamlText), index)
}

func AvpDataFromYamlStr(yamlData string, index int) error {
	node := yaml.Node{}
	err := yaml.Unmarshal([]byte(yamlData), &node)
	if err != nil {
		return err
	}

	for i := 0; i < len(node.Content[0].Content); i += 2 {
		a := node.Content[0].Content[i]
		v := node.Content[0].Content[i+1]
		if err := avpValueFromYamlNode(a, v, index, nil); err != nil {
			return err
		}
	}

	return nil
}

func avpValueFromYamlNode(avp, value *yaml.Node, index int, store *AvpDataStore) error {
	avp_name := avp.Value
	var avp_value any
	if err := value.Decode(&avp_value); err != nil {
		return err
	}
	if avp_value == nil {
		return &ErrInvalidYamlValue{value.Line, value.Content}
	}

	switch value.Kind {
	case yaml.ScalarNode:
		if err := StoreAvpValue(avp_name, &avp_value, index, store); err != nil {
			return err
		}
	case yaml.SequenceNode:
		for _, node := range value.Content {
			if err := avpValueFromYamlNode(avp, node, index, store); err != nil {
				return err
			}
		}
	case yaml.MappingNode:
		s := new(AvpDataStore)
		*s = make(AvpDataStore)
		for i := 0; i < len(value.Content); i += 2 {
			m := value.Content[i]
			v := value.Content[i+1]
			if err := avpValueFromYamlNode(m, v, index, s); err != nil {
				return err
			}
		}
		avp_value = s
		if err := StoreAvpValue(avp_name, &avp_value, index, store); err != nil {
			return err
		}
	default:
		return &ErrInvalidYamlValue{value.Line, value.Content}
	}

	return nil
}

func StoreAvpValue(avpId any, value *any, index int, store *AvpDataStore) error {
	avp, err := Dict.GetAvp(avpId)
	if err != nil {
		return err
	}

	convNumber(value, avp)
	v, err := avp.MakeValue(value)
	if err != nil {
		return err
	}

	if store == nil {
		store = &globalAvpDataStore
	}

	switch i := index; {
	case i == AVP_DATA_APPEND:
		(*store)[avp.Code] = append((*store)[avp.Code], v)
	case i >= 0 && i < len((*store)[avp.Code]):
		(*store)[avp.Code][index] = v
	default:
		return &ErrIndexOutOfRange{avp, index}
	}

	return nil
}

func DelAvpValue(avpId any, index int, store *AvpDataStore) error {
	avp, err := Dict.GetAvp(avpId)
	if err != nil {
		return err
	}

	if store == nil {
		store = &globalAvpDataStore
	}

	if index >= len((*store)[avp.Code]) {
		return &ErrIndexOutOfRange{avp, index}
	}

	if index == AVP_DATA_CLEANUP {
		(*store)[avp.Code] = nil
	} else {
		(*store)[avp.Code] = slices.Delete((*store)[avp.Code], index, index+1)
	}

	return nil
}

// -- Iterators
// --
func AvpDataIter(store *AvpDataStore) iter.Seq[[]*AvpData] {
	if store == nil {
		store = &globalAvpDataStore
	}
	return func(yield func([]*AvpData) bool) {
		for _, value := range *store {
			if !yield(value) {
				break
			}
		}
	}
}

func AvpDataIter2(store *AvpDataStore) iter.Seq2[uint32, []*AvpData] {
	if store == nil {
		store = &globalAvpDataStore
	}
	return func(yield func(uint32, []*AvpData) bool) {
		for code, value := range *store {
			if !yield(code, value) {
				break
			}
		}
	}
}

// -- Data Processing Assistants
// --
func convNumber(n *any, avp *Avp) {
	switch (*n).(type) {
	case int, float64:
	default:
		return
	}

	switch avp.Type {
	case Dict.AvpTypeInteger32():
		*n = int32((*n).(int))
	case Dict.AvpTypeInteger64():
		*n = int64((*n).(int))
	case Dict.AvpTypeUnsigned32():
		*n = uint32((*n).(int))
	case Dict.AvpTypeUnsigned64():
		*n = uint64((*n).(int))
	case Dict.AvpTypeFloat32():
		*n = float32((*n).(float64))
	}
}

func (avp *Avp) enumItemToCode(data any) (int32, error) {
	switch v := data.(type) {
	case int:
		return int32(v), nil
	case string:
		for _, item := range avp.Enum.Items {
			if strings.EqualFold(item.Name, v) {
				return item.Code, nil
			}
		}
	default:
		return 0, &ErrInvalidAvpValue{avp, data}
	}
	return 0, &ErrUnknownEnumItem{Avp: avp, Value: data}
}

func (avp *Avp) encodeAddr(value any) (net.IP, uint32, error) {
	ip := net.ParseIP(value.(string))
	if ip == nil {
		return nil, 0, &ErrInvalidAvpValue{avp, value}
	}

	if ip.To4() != nil {
		return ip, 6, nil
	} else {
		return ip, 18, nil
	}
}

func (avp *Avp) encodeTime(value any) (int64, error) {
	if v, ok := value.(time.Time); ok {
		return v.Unix(), nil
	}

	if v, ok := value.(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.Unix(), nil
		}
	}

	return 0, &ErrInvalidAvpValue{avp, value}
}

func (avp *Avp) encodeOctetString(raw any) ([]byte, error) {
	str := func() string {
		switch (raw).(type) {
		case int:
			return strconv.Itoa((raw).(int))
		default:
			return (raw).(string)
		}
	}()

	// Visited-PLMN-Id
	if avp.Code == 1407 {
		return encodePLMN(str)
	}

	if _, err := strconv.ParseInt(str, 10, 64); err != nil {
		str = hex.EncodeToString([]byte(str))
	}

	bytes := make([]byte, (len(str)+1)/2)
	i := 0
	for ; i < len(str); i++ {
		if i&1 == 0 {
			bytes[i/2] += (str[i] - 0x30)
		} else {
			bytes[i/2] += (str[i] - 0x30) << 4
		}
	}
	if i&1 != 0 {
		bytes[i/2] += 0xf0
	}

	return bytes, nil
}

func encodePLMN(plmn string) ([]byte, error) {
	if len(plmn) != 5 && len(plmn) != 6 {
		return nil, &ErrInvalidValue{plmn}
	}

	bytes := make([]byte, 3)
	bytes[0] = (plmn[0] - 0x30) | (plmn[1]-0x30)<<4
	if len(plmn) == 6 {
		bytes[1] = (plmn[2] - 0x30) | (plmn[3]-0x30)<<4
		bytes[2] = (plmn[4] - 0x30) | (plmn[5]-0x30)<<4
	} else {
		bytes[1] = (plmn[2] - 0x30) | 0xf0
		bytes[2] = (plmn[3] - 0x30) | (plmn[4]-0x30)<<4
	}
	return bytes, nil
}
