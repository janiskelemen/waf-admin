package scheduler

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/janiskelemen/waf-admin/internal/api"
	"github.com/rs/zerolog/log"
)

type Scheduler struct{ jobs []job }
type job struct {
	name, at string
	f        func(context.Context) error
	stop     chan struct{}
}

func New() *Scheduler { return &Scheduler{} }

func (s *Scheduler) AddDaily(name, atHHMM string, f func(context.Context) error) error {
	s.jobs = append(s.jobs, job{name: name, at: atHHMM, f: f, stop: make(chan struct{})})
	return nil
}

func (s *Scheduler) Start() {
	for i := range s.jobs {
		j := s.jobs[i]
		go func() {
			for {
				next := nextAt(j.at)
				d := time.Until(next)
				select {
				case <-time.After(d):
					_ = j.f(context.Background())
				case <-j.stop:
					return
				}
			}
		}()
	}
}

func (s *Scheduler) Stop() {
	for _, j := range s.jobs {
		close(j.stop)
	}
}

func nextAt(hhmm string) time.Time {
	now := time.Now()
	h, m := 3, 30
	parts := strings.Split(hhmm, ":")
	if len(parts) == 2 {
		fmt.Sscanf(parts[0], "%d", &h)
		fmt.Sscanf(parts[1], "%d", &m)
	}
	next := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func RunBackup(ctx context.Context, b api.BackupConfig, c api.CaddyConfig) error {
	ts := time.Now().UTC().Format("20060102-150405")
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for _, p := range []string{c.Caddyfile, c.SitesDir, c.RulesRoot} {
		_ = addToZip(zw, p)
	}
	_ = zw.Close()
	tmp := fmt.Sprintf("/tmp/waf-configs-%s.zip", ts)
	if err := os.WriteFile(tmp, buf.Bytes(), 0o600); err != nil {
		return err
	}
	defer os.Remove(tmp)

	args := []string{"--endpoint-url", b.S3.Endpoint, "s3", "cp", tmp, fmt.Sprintf("s3://%s/%s%s", b.S3.Bucket, strings.TrimPrefix(b.S3.Prefix, "/"), filepath.Base(tmp)), "--acl", "private"}
	cmd := exec.CommandContext(ctx, "aws", args...)
	cmd.Env = append(os.Environ(), "AWS_ACCESS_KEY_ID="+b.S3.AccessKey, "AWS_SECRET_ACCESS_KEY="+b.S3.SecretKey, "AWS_DEFAULT_REGION="+b.S3.Region)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("out", string(out)).Msg("backup upload failed")
		return err
	}
	log.Info().Msg("backup uploaded")
	return nil
}

func addToZip(zw *zip.Writer, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if info.IsDir() {
		return filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if fi.IsDir() {
				return nil
			}
			return zipAddFile(zw, p)
		})
	}
	return zipAddFile(zw, path)
}

func zipAddFile(zw *zip.Writer, p string) error {
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	w, err := zw.Create(p)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}
