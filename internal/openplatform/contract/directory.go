package contract

import (
	"context"
	"net/url"
	"strconv"

	"cn.qfei/contract-cli/internal/openplatform"
)

type DirectoryListInput struct {
	Name         string
	DepartmentID string
	PageSize     int
	PageToken    string
}

func (s *Service) ListEmployees(ctx context.Context, requestContext openplatform.RequestContext, input DirectoryListInput) (openplatform.Response, error) {
	return s.doForCurrentUser(ctx, requestContext, "get-employees", nil, directoryListQuery(input), nil)
}

func (s *Service) ListDepartments(ctx context.Context, requestContext openplatform.RequestContext, input DirectoryListInput) (openplatform.Response, error) {
	return s.doForCurrentUser(ctx, requestContext, "get-departments", nil, directoryListQuery(input), nil)
}

func directoryListQuery(input DirectoryListInput) url.Values {
	query := url.Values{"page_size": {strconv.Itoa(input.PageSize)}}
	if input.Name != "" {
		query.Set("name", input.Name)
	}
	if input.DepartmentID != "" {
		query.Set("department_collection", input.DepartmentID)
	}
	if input.PageToken != "" {
		query.Set("page_token", input.PageToken)
	}
	return query
}
