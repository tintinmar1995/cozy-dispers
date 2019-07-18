package prefixer

// Prefixer interface describes a handle for a specific instance by its domain
// and a specific and unique prefix.
type Prefixer interface {
	DBPrefix() string
	DomainName() string
}

type prefixer struct {
	domain string
	prefix string
}

// UnknownDomainName represents the human-readable string of an empty domain
// name of a prefixer struct
const UnknownDomainName string = "<unknown>"

func (p *prefixer) DBPrefix() string { return p.prefix }

func (p *prefixer) DomainName() string {
	if p.domain == "" {
		return UnknownDomainName
	}
	return p.domain
}

// NewPrefixer returns a prefixer with the specified domain and prefix values.
func NewPrefixer(domain, prefix string) Prefixer {
	return &prefixer{
		domain: domain,
		prefix: prefix,
	}
}

// GlobalPrefixer returns a global prefixer with the wildcard '*' as prefix.
var GlobalPrefixer = NewPrefixer("", "global")

// DataAggregatorPrefixer returns a global prefixer with the wildcard '*' as prefix.
var DataAggregatorPrefixer = NewPrefixer("", "dataaggregator")

// TestDataAggregatorPrefixer returns a global prefixer with the wildcard '*' as prefix.
var TestDataAggregatorPrefixer = NewPrefixer("", "testdataaggregator")

// ConductorPrefixer returns a Conductor prefixer with the wildcard '*' as prefix.
var ConductorPrefixer = NewPrefixer("", "conductor")

// ConceptIndexorPrefixer returns a Concept Indexor prefixer with the wildcard '*' as prefix.
var ConceptIndexorPrefixer = NewPrefixer("", "conceptindexor")

// TestConceptIndexorPrefixer returns a Concept Indexor prefixer for tests with the wildcard '*' as prefix.
var TestConceptIndexorPrefixer = NewPrefixer("", "testconceptindexor")
