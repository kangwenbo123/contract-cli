package cli

import (
	"context"
	"net/http"

	"cn.qfei/contract-cli/internal/invocation"
)

func (a *App) beforeOpenPlatformRequest(ctx context.Context, request *http.Request) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	report := a.inspectEnvironment(ctx, invocation.DefaultMaxDepth)
	// Only the inspection's own deadline is fail-open. Respect cancellation of
	// the original business request instead of sending after the user cancels.
	if err := ctx.Err(); err != nil {
		return err
	}
	invocation.ApplyHeaders(request.Header, report)
	a.logger.Info(
		"open platform invocation environment detected",
		"channel_type", report.ChannelType,
		"agent_source_type", report.AgentSourceType,
		"product_code", report.ProductCode,
		"evidence_type", report.EvidenceType,
		"confidence", report.Confidence,
		"detector_version", report.DetectorVersion,
		"rule_id", report.RuleID,
		"reason", report.Reason,
	)
	return nil
}
