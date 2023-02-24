package count_example

import "gitee.com/aurora-engine/gobatis/examples/count_example/model"

type CountMapper struct {
	// @param
	// @return
	Select func() ([]model.Student, int64, error)
}
