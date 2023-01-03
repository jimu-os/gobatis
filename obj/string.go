package obj

import "time"

type String struct {
	V string
}

func (s *String) Scan(data any) error {
	if data != nil {
		switch data.(type) {
		case time.Time:
			s.V = data.(time.Time).Format("2006-01-02 15:04:05")
		case string:
			s.V = data.(string)
		case []byte:
			s.V = string(data.([]byte))
		}
	}
	return nil
}
