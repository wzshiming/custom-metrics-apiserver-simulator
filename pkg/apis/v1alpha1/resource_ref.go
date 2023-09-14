package v1alpha1

// ResourceRef specifies the kind and version of the resource.
type ResourceRef struct {
	// APIGroup of the referent.
	// +default="v1"
	// +kubebuilder:default="v1"
	APIGroup string `json:"apiGroup,omitempty"`
	// Kind of the referent.
	Kind string `json:"kind"`
}
