package emb

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
)

func ParseUI(web embed.FS, subName string, funcs template.FuncMap, asTemplate bool) (sub fs.FS, t *template.Template, err error) {
	sub = web
	subName = strings.Trim(subName, "/")
	if subName != "" {
		if sub, err = fs.Sub(web, subName); err != nil {
			err = fmt.Errorf("sub %s: %w", subName, err)
			return
		}
	}

	if asTemplate {
		t, err = template.New("").Funcs(funcs).ParseFS(sub, "*.html")
		if err != nil {
			err = fmt.Errorf("parse web template: %w", err)
		}
	}

	return
}
