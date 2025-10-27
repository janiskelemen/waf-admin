package reload

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
)

type Reloader interface {
	Reload(ctx context.Context) error
}

type CaddyAdmin struct {
	sock string
	cfg  string
}

func NewCaddyAdmin(adminSock, caddyfile string) *CaddyAdmin {
	return &CaddyAdmin{sock: adminSock, cfg: caddyfile}
}

func (c *CaddyAdmin) Reload(ctx context.Context) error {
	body, err := os.ReadFile(c.cfg)
	if err != nil {
		return err
	}
	client := &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) { return net.Dial("unix", c.sock) }}}
	req, err := http.NewRequestWithContext(ctx, "POST", "http://unix/load", io.NopCloser(bytesReader(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/caddyfile")
	req.Header.Set("Cache-Control", "must-revalidate")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(resp.Body)
		return &ReloadError{Status: resp.StatusCode, Body: string(b)}
	}
	return nil
}

type ReloadError struct {
	Status int
	Body   string
}

func (e *ReloadError) Error() string { return "caddy reload failed" }

func bytesReader(b []byte) *byteReader { return &byteReader{b: b} }

type byteReader struct{ b []byte }

func (r *byteReader) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}
func (r *byteReader) Close() error { return nil }
