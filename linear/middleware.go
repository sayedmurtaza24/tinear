package linearClient

import (
	"context"
	"net/http"

	"github.com/Yamashou/gqlgenc/clientv2"
)

func GetAuthMiddleware(apiKey string) clientv2.RequestInterceptor {
	return func(
		ctx context.Context,
		req *http.Request,
		gqlInfo *clientv2.GQLRequestInfo,
		res interface{},
		next clientv2.RequestInterceptorFunc,
	) error {
		req.Header.Add("Authorization", apiKey)

		if next != nil {
			return next(ctx, req, gqlInfo, res)
		}

		return nil
	}
}
