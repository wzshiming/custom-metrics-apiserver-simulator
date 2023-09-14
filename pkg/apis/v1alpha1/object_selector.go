package v1alpha1

// ObjectSelector holds information how to match based on namespace and name.
type ObjectSelector struct {
	// MatchNamespaces is a list of namespaces to match.
	// if not set, all namespaces will be matched.
	MatchNamespaces []string `json:"matchNamespaces,omitempty"`
	// MatchNames is a list of names to match.
	// if not set, all names will be matched.
	MatchNames []string `json:"matchNames,omitempty"`
}
