package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	conf "gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/config"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/rpc/parser"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util/format"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util/stringx"
)

const etcTemplate = `Name: {{.serviceName}}.rpc
ListenOn: 127.0.0.1:8080
`

// GenEtc generates the yaml configuration file of the rpc service,
// including host, port monitoring configuration items and etcd configuration
func (g *DefaultGenerator) GenEtc(ctx DirContext, _ parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetEtc()
	etcFilename, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	fileName := filepath.Join(dir.Filename, fmt.Sprintf("%v.yaml", etcFilename))

	text, err := util.LoadTemplate(category, etcTemplateFileFile, etcTemplate)
	if err != nil {
		return err
	}

	return util.With("etc").Parse(text).SaveTo(map[string]interface{}{
		"serviceName": strings.ToLower(stringx.From(ctx.GetServiceName().Source()).ToCamel()),
	}, fileName, false)
}
