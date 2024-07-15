package client

import (
	linearClient "github.com/sayedmurtaza24/tinear/linear"
)

type Client struct {
	client linearClient.LinearClient
}

func New(client linearClient.LinearClient) *Client {
	return &Client{client: client}
}

type nextPageGetter[T any] interface {
	GetHasNextPage() bool
	GetEndCursor() *string
}

type Command[T any] struct {
	Result T
}

type Resumable[T any] struct {
	After  *string
	Result T
}

func first() *int64 {
	var f int64 = 50
	return &f
}

func response[T any](result T) Command[T] {
	return Command[T]{Result: result}
}

func paginated[T any](
	result T,
	pageInfo nextPageGetter[T],
) Resumable[T] {
	cmd := Resumable[T]{
		Result: result,
	}
	if pageInfo.GetHasNextPage() {
		cmd.After = pageInfo.GetEndCursor()
	}
	return cmd
}
