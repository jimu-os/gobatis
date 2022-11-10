package sgo

import (
	"bytes"
	"fmt"
	"strings"
)

type Politic interface {
	// ForEach value 待处理迭代的数据 ctx 上下文数据 item 上下文数据key序列
	ForEach(value any, template string) (string, error)
}

type Combine struct {
	Value    any
	Template string
	Politic
}

func (c Combine) ForEach() (string, error) {
	return c.Politic.ForEach(c.Value, c.Template)
}

// AnalysisForTemplate 解析 for 标签的 文本模板
// template for标签下的文本内容
// ctx 并不是全局的上下文数据，如果 for循环的 item是个 obj ，则ctx将表示 obj
// v 如果 for循环的 item是个 基本类型 v 将代表它
func AnalysisForTemplate(template string, ctx map[string]any, v any) (string, error) {
	buf := bytes.Buffer{}
	template = strings.TrimSpace(template)
	templateByte := []byte(template)
	starIndex := 0
	var item any
	var err error
	for i := starIndex; i < len(templateByte); {
		if templateByte[i] == '{' {
			starIndex = i
			endIndex := i
			for j := starIndex; j < len(templateByte); j++ {
				if templateByte[j] == '}' {
					endIndex = j
					break
				}
			}
			s := template[starIndex+1 : endIndex]
			split := strings.Split(s, ".")
			if len(split) > 1 && ctx != nil {
				item, err = ctxValue(ctx, split)
				if err != nil {
					return "", fmt.Errorf("%s,'%s' not found", template, s)
				}
			} else {
				item = v
			}
			if item == nil {
				return "", fmt.Errorf("%s,'%s' not found", template, s)
			}
			switch item.(type) {
			case string:
				buf.WriteString(fmt.Sprintf(" '%s' ", item.(string)))
			case int:
				buf.WriteString(fmt.Sprintf(" %d ", item.(int)))
			case float64:
				buf.WriteString(fmt.Sprintf(" %f ", item.(float64)))
			default:
				// 其他复杂数据类型
				if handle := dataHandle(item); handle != "" {
					buf.WriteString(" " + handle + " ")
				}
			}
			i = endIndex + 1
			continue
		}
		buf.WriteByte(templateByte[i])
		i++
	}
	return buf.String(), nil
}
