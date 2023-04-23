package mapper_example

import "gitee.com/aurora-engine/gobatis/examples/mapper_example/model"

type StudentMapper struct {
	AddOne func(student model.Student) error
}
