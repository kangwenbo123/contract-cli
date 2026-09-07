package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	"cn.qfei/contract-cli/internal/openplatform"
)

func (a *App) renderContractMCPResponse(options commandOptions, response openplatform.Response) error {
	responseErr := validateContractMCPResponse(response.Body)
	if responseErr != nil {
		a.logger.Error("contract MCP response rejected", "error", responseErr.Error())
	}
	// Preserve the response for scripts even when the command must exit with an error.
	outputErr := a.renderOpenPlatformResponse(options, response)
	return errors.Join(responseErr, outputErr)
}

func validateContractMCPResponse(body []byte) error {
	var envelope struct {
		Code    *int64 `json:"code"`
		Success *bool  `json:"success"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return errors.New("invalid contract MCP response envelope")
	}
	if envelope.Code == nil {
		return errors.New("contract MCP response is missing a business code")
	}
	if *envelope.Code != 0 || (envelope.Success != nil && !*envelope.Success) {
		return fmt.Errorf("contract MCP request failed with business code %d; see response output for details", *envelope.Code)
	}
	return nil
}
