package invocation

import "net/http"

const (
	HeaderChannelType     = "X-Qfei-Channel-Type"
	HeaderAgentSourceType = "X-Qfei-Agent-Source-Type"
	HeaderProductCode     = "X-Qfei-Product-Code"
	HeaderEvidenceType    = "X-Qfei-Evidence-Type"
	HeaderConfidence      = "X-Qfei-Channel-Confidence"
	HeaderDetectorVersion = "X-Qfei-Detector-Version"
	HeaderRuleID          = "X-Qfei-Rule-Id"

	ProductCodeContract = "contract"
	legacyRequestSource = "X-Qfei-Request-Source-Type"
)

func ApplyHeaders(headers http.Header, result Result) {
	// Remove the superseded field even when a caller supplied it explicitly.
	headers.Del(legacyRequestSource)
	setOrDeleteHeader(headers, HeaderChannelType, result.ChannelType)
	setOrDeleteHeader(headers, HeaderAgentSourceType, result.AgentSourceType)
	setOrDeleteHeader(headers, HeaderProductCode, result.ProductCode)
	setOrDeleteHeader(headers, HeaderEvidenceType, result.EvidenceType)
	setOrDeleteHeader(headers, HeaderConfidence, result.Confidence)
	setOrDeleteHeader(headers, HeaderDetectorVersion, result.DetectorVersion)
	setOrDeleteHeader(headers, HeaderRuleID, result.RuleID)
}

func setOrDeleteHeader(headers http.Header, name, value string) {
	if value == "" {
		headers.Del(name)
		return
	}
	headers.Set(name, value)
}
