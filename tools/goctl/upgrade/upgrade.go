package upgrade

import (
	"fmt"

	"github.com/urfave/cli"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/rpc/execx"
)

// Upgrade gets the latest goctl by
// go get -u gitlab.deepwisdomai.com/infra/go-zero/tools/goctl
func Upgrade(_ *cli.Context) error {
	info, err := execx.Run("GO111MODULE=on GOPROXY=https://goproxy.cn/,direct go get -u gitlab.deepwisdomai.com/infra/go-zero/tools/goctl", "")
	if err != nil {
		return err
	}

	fmt.Print(info)
	return nil
}
