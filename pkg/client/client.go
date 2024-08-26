package client

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Yamashou/gqlgenc/clientv2"
	linearClient "github.com/sayedmurtaza24/tinear/linear"
)

const linearBaseUrL = "https://api.linear.app/graphql"

type Client struct {
	client    linearClient.LinearClient
	rawClient *clientv2.Client
}

func getApiKey() string {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY is not set")
	}
	return apiKey
}

func initLinearClient() linearClient.LinearClient {
	md := linearClient.GetAuthMiddleware(getApiKey())

	client := linearClient.NewClient(http.DefaultClient, linearBaseUrL, nil, md)

	return client
}

func initLinearRawClient() *clientv2.Client {
	md := linearClient.GetAuthMiddleware(getApiKey())

	client := clientv2.NewClient(http.DefaultClient, linearBaseUrL, nil, md)

	client.CustomDo = func(
		ctx context.Context,
		req *http.Request,
		gqlInfo *clientv2.GQLRequestInfo,
		res interface{},
	) error {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.Header.Get("Content-Encoding") == "gzip" {
			resp.Body, err = gzip.NewReader(resp.Body)
			if err != nil {
				return fmt.Errorf("gzip decode failed: %w", err)
			}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		err = json.Unmarshal(body, res)
		if err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	}

	return client
}

func New() *Client {
	return &Client{
		client:    initLinearClient(),
		rawClient: initLinearRawClient(),
	}
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
