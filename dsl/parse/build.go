package parse

import (
	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/internal/graphbuild"
	"github.com/trustmaster/goflow/dsl/types"
)

// Build transforms a parsed Definition into a runnable *goflow.Graph.
func Build(def *types.Definition, f *goflow.Factory) (*goflow.Graph, error) {
	return graphbuild.Build(def, f)
}
