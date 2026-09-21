package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"cn.qfei/contract-cli/internal/openplatform"
	contractsvc "cn.qfei/contract-cli/internal/openplatform/contract"
)

func (a *App) runDirectory(ctx context.Context, resource string, args []string) (err error) {
	a.logger.Info("directory query started", "resource", resource)
	defer func() {
		if err != nil {
			a.logger.Error("directory query failed", "resource", resource, "error_type", fmt.Sprintf("%T", err))
		}
	}()
	if len(args) == 0 || args[0] != "list" {
		return fmt.Errorf("usage: contract-cli %s list [flags]", resource)
	}
	parsed, err := parseDirectoryArgs(resource, args[1:])
	if err != nil {
		return err
	}
	input, err := directoryListInput(parsed)
	if err != nil {
		return err
	}
	a.logger.Info("directory query scope", "resource", resource, "by_name", input.Name != "", "by_department", input.DepartmentID != "", "page_size", input.PageSize, "has_page_token", input.PageToken != "")
	options := parseCommandOptions(parsed)
	response, err := a.queryDirectory(ctx, resource, options, input)
	if err != nil {
		return safeDirectoryRequestError(err)
	}
	return a.renderContractMCPResponse(options, response)
}

type directoryRequestError struct {
	cause   error
	message string
}

func (e *directoryRequestError) Error() string { return e.message }
func (e *directoryRequestError) Unwrap() error { return e.cause }

func safeDirectoryRequestError(cause error) error {
	var traced *openplatform.TraceError
	if !errors.As(cause, &traced) {
		return cause
	}
	detail := "request could not be completed"
	var status *openplatform.HTTPStatusError
	if errors.As(cause, &status) {
		detail = fmt.Sprintf("HTTP status %d", status.StatusCode)
	}
	if errors.Is(cause, ErrDeviceReauthorizationRequired) {
		detail = ErrDeviceReauthorizationRequired.Error() + "; user confirmation is required before checking authorization status and starting or restarting Device authorization"
	}
	return &directoryRequestError{cause: cause, message: fmt.Sprintf("directory request failed: %s (trace_id=%s)", detail, traced.TraceID)}
}

func parseDirectoryArgs(resource string, args []string) (parsedArgs, error) {
	flags := map[string]struct{}{"--profile": {}, "--as": {}, "--output": {}, "--name": {}, "--department-id": {}, "--page-size": {}, "--page-token": {}}
	parsed, err := parseArgs(args, flags, commonBoolFlags())
	if err != nil {
		return parsedArgs{}, err
	}
	if len(parsed.positionals) != 0 {
		return parsedArgs{}, fmt.Errorf("usage: contract-cli %s list [flags]", resource)
	}
	if err := validateDirectorySelectors(resource, parsed); err != nil {
		return parsedArgs{}, err
	}
	return parsed, nil
}

func validateDirectorySelectors(resource string, parsed parsedArgs) error {
	for flag, values := range parsed.values {
		if len(values) > 1 {
			return fmt.Errorf("%s may only be specified once", flag)
		}
	}
	for _, flag := range []string{"--name", "--department-id"} {
		if parsed.HasValue(flag) && strings.TrimSpace(parsed.String(flag)) == "" {
			return fmt.Errorf("%s must not be empty", flag)
		}
	}
	if parsed.HasValue("--name") && parsed.HasValue("--department-id") {
		return fmt.Errorf("--name and --department-id are mutually exclusive")
	}
	if resource == "employee" && !parsed.HasValue("--name") && !parsed.HasValue("--department-id") {
		return fmt.Errorf("employee list requires --name or --department-id")
	}
	return nil
}

func directoryListInput(parsed parsedArgs) (contractsvc.DirectoryListInput, error) {
	pageSize := 10
	if parsed.HasValue("--page-size") {
		value, err := strconv.Atoi(parsed.String("--page-size"))
		if err != nil || value < 1 || value > 200 {
			return contractsvc.DirectoryListInput{}, fmt.Errorf("--page-size must be an integer between 1 and 200")
		}
		pageSize = value
	}
	return contractsvc.DirectoryListInput{Name: parsed.String("--name"), DepartmentID: parsed.String("--department-id"), PageSize: pageSize, PageToken: parsed.String("--page-token")}, nil
}

func (a *App) queryDirectory(ctx context.Context, resource string, options commandOptions, input contractsvc.DirectoryListInput) (openplatform.Response, error) {
	client, requestContext, err := a.openPlatformClientAndContextForOptions(options, contractMCPPathPrefix+"/"+resource+"s", openplatform.IdentityPolicyUserOnly)
	if err != nil {
		return openplatform.Response{}, err
	}
	service := contractsvc.NewService(client)
	if resource == "employee" {
		return service.ListEmployees(ctx, requestContext, input)
	}
	return service.ListDepartments(ctx, requestContext, input)
}
