package new

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/urfave/cli"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/api/gogen"
	conf "gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/config"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util"
)

const apiTemplate = `
type Request {
  Name string ` + "`" + `path:"name,options=you|me"` + "`" + ` 
}

type Response {
  Message string ` + "`" + `json:"message"` + "`" + `
}

service {{.name}}-api {
  @handler {{.handler}}Handler
  get /from/:name(Request) returns (Response);
}
`

// CreateServiceCommand fast create service
func CreateServiceCommand(c *cli.Context) error {
	args := c.Args()
	dirName := args.First()
	if len(dirName) == 0 {
		dirName = "greet"
	}

	if strings.Contains(dirName, "-") {
		return errors.New("api new command service name not support strikethrough, because this will used by function name")
	}

	abs, err := filepath.Abs(dirName)
	if err != nil {
		return err
	}

	err = util.MkdirIfNotExist(abs)
	if err != nil {
		return err
	}

	dirName = filepath.Base(filepath.Clean(abs))
	filename := dirName + ".api"
	apiFilePath := filepath.Join(abs, filename)
	fp, err := os.Create(apiFilePath)
	if err != nil {
		return err
	}

	defer fp.Close()
	t := template.Must(template.New("template").Parse(apiTemplate))
	if err := t.Execute(fp, map[string]string{
		"name":    dirName,
		"handler": strings.Title(dirName),
	}); err != nil {
		return err
	}

	err = gogen.DoGenProject(apiFilePath, abs, conf.DefaultFormat)
	return err
}
