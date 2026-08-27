package invocation

import "net/http"

const (
	HeaderRequestSourceType = "X-Qfei-Request-Source-Type"
	HeaderChannelType       = "X-Qfei-Channel-Type"
	HeaderEvidenceType      = "X-Qfei-Evidence-Type"
	HeaderConfidence        = "X-Qfei-Channel-Confidence"
	HeaderDetectorVersion   = "X-Qfei-Detector-Version"
	HeaderRuleID            = "X-Qfei-Rule-Id"
)

func ApplyHeaders(headers http.Header, result Result) {
	setOrDeleteHeader(headers, HeaderRequestSourceType, result.RequestSourceType)
	setOrDeleteHeader(headers, HeaderChannelType, result.ChannelType)
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
