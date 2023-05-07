package gobatis

import (
	"fmt"
	"github.com/beevik/etree"
)

func ValueTag(element *etree.Element, template string, ctx map[string]any) (string, string, []any, error) {
	//获取标签
	tag := element.Tag
	analysisTemplate, t, param, err := AnalysisTemplate(template, ctx)
	if err != nil {
		return "", "", nil, nil
	}
	analysisTemplate = fmt.Sprintf("%s %s", tag, analysisTemplate)
	t = fmt.Sprintf("%s %s", tag, t)
	return analysisTemplate, t, param, nil
}
