package cli

import (
	"context"
	"fmt"

	"cn.qfei/contract-cli/internal/openplatform"
	contractsvc "cn.qfei/contract-cli/internal/openplatform/contract"
)

func (a *App) runContractSearchFields(ctx context.Context, args []string) (err error) {
	a.logger.Info("list contract search filter fields")
	defer func() {
		if err != nil {
			a.logger.Error("list contract search filter fields failed", "error", err.Error())
		}
	}()

	parsed, err := parseArgs(args, commonValueFlags("--keyword"), commonBoolFlags())
	if err != nil {
		return err
	}
	if len(parsed.positionals) != 0 {
		return fmt.Errorf("usage: contract-cli contract search-fields [--keyword <field-name>] [flags]")
	}
	if parsed.HasValue("--input-file") || parsed.HasValue("--data") {
		return fmt.Errorf("contract search-fields does not accept --input-file or --data")
	}
	options := parseCommandOptions(parsed)
	keyword := parsed.String("--keyword")
	a.logger.Info("resolve contract search field scope", "has_keyword", keyword != "")
	client, requestContext, err := a.openPlatformClientAndContextForOptions(options,
		contractMCPPathPrefix+"/contracts/search/filter_fields", openplatform.IdentityPolicyUserOnly)
	if err != nil {
		return err
	}
	response, err := contractsvc.NewService(client).ListSearchFilterFields(ctx, requestContext, keyword)
	if err != nil {
		return err
	}
	return a.renderContractMCPResponse(options, response)
}
