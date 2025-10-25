package render

import "context"

type Driver interface {
	LayoutSites() string
	LayoutRulesRoot() string
	Validate(ctx context.Context) error
}
