package gobatis

import (
	"bytes"
	"fmt"
	"strings"
)

// Politic for 标签迭代实现接口扩展 标准切片之外的 List 数据支持
type Politic interface {
	// ForEach value 待处理迭代的数据 ctx 上下文数据 item 上下文数据key序列
	ForEach(value any, template string, separator string) (string, string, []any, error)
}

type Combine struct {
	Value     any
	Template  string
	Separator string
	Politic
}

func (c Combine) ForEach() (string, string, []any, error) {
	return c.Politic.ForEach(c.Value, c.Template, c.Separator)
}

// AnalysisForTemplate 解析 for 标签的 文本模板
// template for标签下的文本内容
// ctx 并不是全局的上下文数据，如果 for循环的 item是个 obj ，则ctx将表示 obj
// v 如果 for循环的 item是个 基本类型 v 将代表它
func AnalysisForTemplate(template string, ctx map[string]any, v any) (string, string, []any, error) {
	buf := bytes.Buffer{}
	templateBuf := bytes.Buffer{}
	params := []any{}
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
				item, err = sliceCtxValue(ctx, split)
				if err != nil {
					return "", "", nil, fmt.Errorf("%s,'%s' not found", template, s)
				}
			} else {
				item = v
			}
			if item == nil {
				return "", "", nil, fmt.Errorf("%s,'%s' not found", template, s)
			}
			/*switch item.(type) {
			case string:
				buf.WriteString(fmt.Sprintf(" '%s' ", item.(string)))
				templateBuf.WriteString("?")
				params = append(params, item)
			case int:
				buf.WriteString(fmt.Sprintf(" %d ", item.(int)))
				templateBuf.WriteString("?")
				params = append(params, item)
			case float64:
				buf.WriteString(fmt.Sprintf(" %f ", item.(float64)))
				templateBuf.WriteString("?")
				params = append(params, item)
			default:
				// 其他复杂数据类型
				if handle, e := dataHandle(item); e != nil {
					return "", "", nil, e
				} else {
					var v string
					switch handle.(type) {
					case string:
						v = "'" + handle.(string) + "'"
					case int:
						v = strconv.Itoa(handle.(int))
					case float64:
						v = strconv.FormatFloat(handle.(float64), 'f', 'f', 64)
					case bool:
						v = strconv.FormatBool(handle.(bool))
					}
					buf.WriteString(v)
					templateBuf.WriteString("?")
					params = append(params, handle)
				}
			}*/
			value, err := elementValue(item)
			if err != nil {
				return "", "", nil, err
			}
			buf.WriteString(" " + value + " ")
			templateBuf.WriteString("?")
			params = append(params, item)
			i = endIndex + 1
			continue
		}
		buf.WriteByte(templateByte[i])
		templateBuf.WriteByte(templateByte[i])
		i++
	}
	return buf.String(), templateBuf.String(), params, nil
}
