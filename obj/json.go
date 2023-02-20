package obj

import "encoding/json"

// Json MySQL Json 数据结构
type Json struct {
	V any
}

func (receiver *Json) Scan(data any) error {
	if data == nil {
		return nil
	}
	switch data.(type) {
	case string:
		var value any
		v := data.(string)
		err := json.Unmarshal([]byte(v), &value)
		if err != nil {
			return err
		}
		receiver.V = value
	}
	return nil
}
