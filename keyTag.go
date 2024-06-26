package gobatis

import (
	"fmt"
	"github.com/beevik/etree"
)

func keyTag(element *etree.Element, template string, ctx map[string]any) (string, string, []any, error) {
	tag := element.Tag
	analysisTemplate, t, param, err := AnalysisTemplate(template, ctx)
	if err != nil {
		return "", "", nil, nil
	}
	if analysisTemplate != "" {
		analysisTemplate = fmt.Sprintf("%s %s", tag, analysisTemplate)
		t = fmt.Sprintf("%s %s", tag, t)
	} else {
		analysisTemplate = tag
		t = tag
	}
	return analysisTemplate, t, param, nil
}
