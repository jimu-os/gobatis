package count_example

import (
	"gitee.com/aurora-engine/gobatis/examples/count_example/model"
	"gitee.com/aurora-engine/gobatis/opt"
)

type CountMapper struct {
	// @param
	// @return
	Select func(...opt.Opt) ([]model.Student, int64, error)

	SelectTest func(any, ...opt.Opt) ([]model.Student, int64, error)
}
