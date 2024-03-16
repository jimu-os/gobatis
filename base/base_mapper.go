package base

type Query map[string]any

func (query Query) Eq() {

}

type Mapper[T any] struct {
	Insert func(T)

	Delete func(T)

	Update func(T)

	Select func(T)

	SelectById func(any) T

	SelectByMap func(query Query) []T

	SelectOne func(T) T

	SelectList func(T) []T
}
