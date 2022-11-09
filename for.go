package sgo

type Politic interface {
	// ForEach value 待处理迭代的数据 ctx 上下文数据 item 上下文数据key序列
	ForEach(value any, ctx map[string]any, item []string) (string, error)
}

type Combine struct {
	Value any
	Ctx   map[string]any
	Keys  []string
	Politic
}

func (c Combine) ForEach() (string, error) {
	return c.Politic.ForEach(c.Value, c.Ctx, c.Keys)
}
