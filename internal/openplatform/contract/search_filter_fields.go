package contract

import (
	"context"
	"net/url"

	"cn.qfei/contract-cli/internal/openplatform"
)

func (s *Service) ListSearchFilterFields(ctx context.Context, requestContext openplatform.RequestContext, keyword string) (openplatform.Response, error) {
	query := url.Values{}
	if keyword != "" {
		query.Set("keyword", keyword)
	}
	return s.doForCurrentUser(ctx, requestContext, "list-contract-search-filter-fields", nil, query, nil)
}
