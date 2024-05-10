package resumable

type Command[T any] struct {
	After  *string
	Result T
}

type nextPageGetter[T any] interface {
	GetHasNextPage() bool
	GetEndCursor() *string
}

func FromLinearClientResponse[T any](
	result T,
	pageInfo nextPageGetter[T],
) Command[T] {
	cmd := Command[T]{
		Result: result,
	}
	if pageInfo.GetHasNextPage() {
		cmd.After = pageInfo.GetEndCursor()
	}
	return cmd
}
