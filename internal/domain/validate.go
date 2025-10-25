package domain

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/janiskelemen/waf-admin/internal/render"
	"github.com/janiskelemen/waf-admin/internal/storage"
)

func ListSites(ctx context.Context, dr render.Driver, st storage.Storage) ([]SiteInfo, error) {
	ents, err := st.List(ctx, dr.LayoutSites())
	if err != nil {
		return nil, err
	}
	var out []SiteInfo
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".caddy") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".caddy")
		info := SiteInfo{
			Name:        name,
			SnippetPath: filepath.Join(dr.LayoutSites(), e.Name()),
			RulesPath:   filepath.Join(dr.LayoutRulesRoot(), name, "rules"),
			HasSnippet:  true,
		}
		if _, err := st.List(ctx, info.RulesPath); err == nil {
			info.HasRulesDir = true
		}
		out = append(out, info)
	}
	return out, nil
}
