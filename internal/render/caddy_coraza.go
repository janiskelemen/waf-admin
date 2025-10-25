package render

import (
	"context"
	"os/exec"
)

type CaddyOptions struct {
	AdminSocket string
	Caddyfile   string
	SitesDir    string
	RulesRoot   string
}

type CaddyCoraza struct{ CaddyOptions }

func NewCaddyCoraza(o CaddyOptions) *CaddyCoraza { return &CaddyCoraza{CaddyOptions:o} }

func (c *CaddyCoraza) LayoutSites() string     { return c.SitesDir }
func (c *CaddyCoraza) LayoutRulesRoot() string { return c.RulesRoot }

func (c *CaddyCoraza) Validate(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "caddy", "validate", "--config", c.Caddyfile)
	return cmd.Run()
}
