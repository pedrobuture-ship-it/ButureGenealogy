package genetics

// Relationship and inbreeding coefficients are NOT computed in this version.
//
// The planned method is Henderson's tabular algorithm for the additive
// relationship matrix, with Wright's coefficient of inbreeding F_x = a_sd / 2
// (half the relationship between the parents). Until that is implemented and
// tested against published reference cases, the application reports
// "not calculated" instead of showing an approximate number.

// Status values for coefficient results.
const (
	StatusNotCalculated = "not_calculated"
	StatusCalculated    = "calculated"
)

// CoefficientResult is the outcome of an inbreeding/relationship request.
type CoefficientResult struct {
	Status  string   `json:"status"`
	Value   *float64 `json:"value"`
	Method  string   `json:"method"`
	Message string   `json:"message"`
}

// RelationshipCalculator is the extension point for the formal engine.
type RelationshipCalculator interface {
	Inbreeding(g Graph, id string) CoefficientResult
	Relationship(g Graph, a, b string) CoefficientResult
}

// UnavailableCalculator is the current implementation: it never invents a value.
type UnavailableCalculator struct{}

const notCalculatedMessage = "não calculado nesta versão"

func (UnavailableCalculator) Inbreeding(Graph, string) CoefficientResult {
	return CoefficientResult{Status: StatusNotCalculated, Message: notCalculatedMessage}
}

func (UnavailableCalculator) Relationship(Graph, string, string) CoefficientResult {
	return CoefficientResult{Status: StatusNotCalculated, Message: notCalculatedMessage}
}
