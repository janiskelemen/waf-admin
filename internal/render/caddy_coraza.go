package render

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

type CaddyOptions struct {
	AdminSocket string
	Caddyfile   string
	SitesDir    string
	RulesRoot   string
}

type CaddyCoraza struct{ CaddyOptions }

func NewCaddyCoraza(o CaddyOptions) *CaddyCoraza { return &CaddyCoraza{CaddyOptions: o} }

func (c *CaddyCoraza) LayoutSites() string     { return c.SitesDir }
func (c *CaddyCoraza) LayoutRulesRoot() string { return c.RulesRoot }

func (c *CaddyCoraza) Validate(ctx context.Context) error {
	body, err := os.ReadFile(c.Caddyfile)
	if err != nil {
		return err
	}

	client := &http.Client{Transport: c.unixTransport()}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://unix/adapt", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/caddyfile")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("caddy adapt failed: %s", bytes.TrimSpace(msg))
	}
	io.Copy(io.Discard, resp.Body)
	return nil
}

func (c *CaddyCoraza) unixTransport() *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", c.AdminSocket)
		},
	}
}
