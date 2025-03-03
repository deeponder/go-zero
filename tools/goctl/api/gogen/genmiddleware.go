package gogen

import (
	"strings"

	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/api/spec"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/config"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util/format"
)

var middlewareImplementCode = `
package middleware

import "net/http"

type {{.name}} struct {
}

func New{{.name}}() *{{.name}} {	
	return &{{.name}}{}
}

func (m *{{.name}})Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation

		// Passthrough to next handler if need 
		next(w, r)
	}	
}
`

func genMiddleware(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	middlewares := getMiddleware(api)
	for _, item := range middlewares {
		middlewareFilename := strings.TrimSuffix(strings.ToLower(item), "middleware") + "_middleware"
		filename, err := format.FileNamingFormat(cfg.NamingFormat, middlewareFilename)
		if err != nil {
			return err
		}

		name := strings.TrimSuffix(item, "Middleware") + "Middleware"
		err = genFile(fileGenConfig{
			dir:             dir,
			subdir:          middlewareDir,
			filename:        filename + ".go",
			templateName:    "contextTemplate",
			builtinTemplate: middlewareImplementCode,
			data: map[string]string{
				"name": strings.Title(name),
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
