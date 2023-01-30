package types

// Metadata specifies plugin metadata
type Metadata struct {
	Name   string `json:"name" yaml:"name"`
	Target string `json:"target" yaml:"target"`
}
