package generator

import (
	"fmt"
	"path/filepath"

	conf "gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/config"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/rpc/parser"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util/format"
)

const svcTemplate = `package svc

import {{.imports}}

type ServiceContext struct {
	Config config.Config
	DynamicConfig config.DynamicConfig
}

func NewServiceContext(c config.Config, dc config.DynamicConfig) *ServiceContext {
	return &ServiceContext{
		Config:c,
		DynamicConfig: dc,
	}
}
`

// GenSvc generates the servicecontext.go file, which is the resource dependency of a service,
// such as rpc dependency, model dependency, etc.
func (g *DefaultGenerator) GenSvc(ctx DirContext, _ parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetSvc()
	svcFilename, err := format.FileNamingFormat(cfg.NamingFormat, "service_context")
	if err != nil {
		return err
	}

	fileName := filepath.Join(dir.Filename, svcFilename+".go")
	text, err := util.LoadTemplate(category, svcTemplateFile, svcTemplate)
	if err != nil {
		return err
	}

	return util.With("svc").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"imports": fmt.Sprintf(`"%v"`, ctx.GetConfig().Package),
	}, fileName, false)
}
