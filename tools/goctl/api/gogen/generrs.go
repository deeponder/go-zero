package gogen

import (
	"fmt"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/api/spec"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/config"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util/format"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/vars"
	"strings"
)

const (
	errsFile     = "errs"
	errsTemplate = `
package types

import (
	{{.imports}}
)

var ErrorAuthCheck = httpx.ErrorJson{
	ErrNo:  100,
	ErrMsg: "check auth fail.",
}

// 4000000-4999999 参数检查错误
var ErrorParamInvalid = httpx.ErrorJson{
	ErrNo:  4000,
	ErrMsg: "param invalid",
}

// 内部逻辑错误
var ErrorSystemError = httpx.ErrorJson{
	ErrNo:  5000,
	ErrMsg: "system internal error",
}

// model层错误
var ErrorDbInsert = httpx.ErrorJson{
	ErrNo:  3101,
	ErrMsg: "db insert error: %s",
}
var ErrorDbUpdate = httpx.ErrorJson{
	ErrNo:  3102,
	ErrMsg: "db update error: %s",
}
var ErrorDbSelect = httpx.ErrorJson{
	ErrNo:  3103,
	ErrMsg: "db get error: %s",
}

// 第三方sdk错误
// redis
var ErrorRedisGet = httpx.ErrorJson{
	ErrNo:  3201,
	ErrMsg: "redis get error: %s",
}
var ErrorRedisSet = httpx.ErrorJson{
	ErrNo:  3202,
	ErrMsg: "redis set error: %s",
}
`
)

func genErrImports() string {
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s/rest/httpx\"", vars.ProjectOpenSourceURL))
	return strings.Join(imports, "\n\t")
}

func genErrs(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	errFilename, err := format.FileNamingFormat(cfg.NamingFormat, errsFile)
	if err != nil {
		return err
	}
	imports := genErrImports()

	errFilename = errFilename + ".go"
	//filename := path.Join(dir, typesDir, errFilename)
	//os.Remove(filename)

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          typesDir,
		filename:        errFilename,
		templateName:    "errsTemplate",
		category:        "",
		templateFile:    "",
		builtinTemplate: errsTemplate,
		data: map[string]interface{}{
			"imports": imports,
		},
	})
}
