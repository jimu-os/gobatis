package obj

import (
	"encoding/json"
)

// Json MySQL Json 数据结构
type Json struct {
	V any
}

func (receiver *Json) Scan(data any) error {
	if data == nil {
		return nil
	}
	var value any
	var v string
	switch data.(type) {
	case []uint8:
		uint8s := data.([]uint8)
		v = string(uint8s)
	case string:
		v = data.(string)
	}
	err := json.Unmarshal([]byte(v), &value)
	if err != nil {
		return err
	}
	receiver.V = value
	return nil
}
