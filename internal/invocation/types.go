package invocation

const (
	DefaultMaxDepth = 32
	DetectorVersion = "process-ancestry-v2"
)

type Process struct {
	Depth      int    `json:"depth"`
	PID        int32  `json:"pid"`
	PPID       int32  `json:"ppid"`
	Name       string `json:"name,omitempty"`
	Executable string `json:"executable,omitempty"`
}

type ApplicationIdentity struct {
	ProcessDepth      int    `json:"process_depth"`
	BundlePath        string `json:"bundle_path,omitempty"`
	BundleID          string `json:"bundle_id,omitempty"`
	TeamID            string `json:"team_id,omitempty"`
	ExecutablePath    string `json:"executable_path,omitempty"`
	PackageFamilyName string `json:"package_family_name,omitempty"`
	Publisher         string `json:"publisher,omitempty"`
	CertificateSHA256 string `json:"certificate_sha256,omitempty"`
	Version           string `json:"version,omitempty"`
	SignatureValid    bool   `json:"signature_valid"`
}

type Result struct {
	RequestSourceType string               `json:"request_source_type"`
	ChannelType       string               `json:"channel_type"`
	EvidenceType      string               `json:"evidence_type"`
	Confidence        string               `json:"confidence"`
	DetectorVersion   string               `json:"detector_version"`
	Platform          string               `json:"platform"`
	RuleID            string               `json:"rule_id,omitempty"`
	Reason            string               `json:"reason"`
	MatchedProcess    *Process             `json:"matched_process,omitempty"`
	Application       *ApplicationIdentity `json:"application,omitempty"`
	Processes         []Process            `json:"processes,omitempty"`
	Warnings          []string             `json:"warnings,omitempty"`
}
