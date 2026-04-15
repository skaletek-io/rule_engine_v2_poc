package rule

const (
	StatusActive   = "active"
	StatusDraft    = "draft"
	StatusDisabled = "disabled"

	ModeLive   = "live"
	ModeShadow = "shadow"

	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)

// Rule is a compiled condition bound to a template.
type Rule struct {
	ID         string
	Name       string
	TemplateID string
	Expression string
	Severity   string // critical | high | medium | low
	Message    string
	Priority   int    // lower = evaluated first
	Status     string // active | draft | disabled
	Mode       string // live | shadow
}
