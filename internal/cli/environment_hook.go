package cli

import (
	"context"
	"net/http"

	"cn.qfei/contract-cli/internal/invocation"
)

func (a *App) beforeOpenPlatformRequest(_ context.Context, request *http.Request) error {
	report := a.inspectEnvironment(invocation.DefaultMaxDepth)
	invocation.ApplyHeaders(request.Header, report)
	a.logger.Info(
		"open platform invocation environment detected",
		"channel_type", report.ChannelType,
		"evidence_type", report.EvidenceType,
		"confidence", report.Confidence,
		"detector_version", report.DetectorVersion,
		"rule_id", report.RuleID,
	)
	return nil
}
