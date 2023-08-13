package mapper_example

import (
	"database/sql"
	"gitee.com/aurora-engine/gobatis/examples/mapper_example/model"
)

type StudentMapper struct {
	AddOne   func(student model.Student) error
	InsertId func(student model.Student) (int64, int64, error)
	Adds     func(ctx any) error

	QueryAll  func(ctx any) ([]model.Student, error)
	QueryPage func() ([]model.Student, int64, error)

	Update func(student model.Student, tx *sql.Tx) (int64, error)

	QueryIf func(any) (model.Student, error)
}
