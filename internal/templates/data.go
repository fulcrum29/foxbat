package templates

import (
	"github.com/prometheus/common/model"
)

// RecommenderData represents the data structure passed to the results.html template.
type RecommenderData struct {
	CPU       model.Vector
	Memory    model.Vector
	Namespace string
	Workload  string
}

// SelectOptionsData represents the data structure passed to the select-options.html template.
type SelectOptionsData struct {
	Items  []string
	Label  string
	Name   string
	Target string
}

// IndexData represents the data structure passed to the index.html template.
type IndexData struct {
	Namespaces []string
}
