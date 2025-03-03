package generator

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.deepwisdomai.com/infra/go-zero/core/logx"
	"gitlab.deepwisdomai.com/infra/go-zero/core/stringx"
	conf "gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/config"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/rpc/execx"
)

var cfg = &conf.Config{
	NamingFormat: "gozero",
}

func TestRpcGenerate(t *testing.T) {
	_ = Clean()
	dispatcher := NewDefaultGenerator()
	err := dispatcher.Prepare()
	if err != nil {
		logx.Error(err)
		return
	}
	projectName := stringx.Rand()
	g := NewRPCGenerator(dispatcher, cfg)

	// case go path
	src := filepath.Join(build.Default.GOPATH, "src")
	_, err = os.Stat(src)
	if err != nil {
		return
	}

	projectDir := filepath.Join(src, projectName)
	srcDir := projectDir
	defer func() {
		_ = os.RemoveAll(srcDir)
	}()
	err = g.Generate("./test.proto", projectDir, []string{src})
	assert.Nil(t, err)
	_, err = execx.Run("go test "+projectName, projectDir)
	if err != nil {
		assert.True(t, func() bool {
			return strings.Contains(err.Error(), "not in GOROOT") || strings.Contains(err.Error(), "cannot find package")
		}())
	}

	// case go mod
	workDir := t.TempDir()
	name := filepath.Base(workDir)
	_, err = execx.Run("go mod init "+name, workDir)
	if err != nil {
		logx.Error(err)
		return
	}

	projectDir = filepath.Join(workDir, projectName)
	err = g.Generate("./test.proto", projectDir, []string{src})
	assert.Nil(t, err)
	_, err = execx.Run("go test "+projectName, projectDir)
	if err != nil {
		assert.True(t, func() bool {
			return strings.Contains(err.Error(), "not in GOROOT") || strings.Contains(err.Error(), "cannot find package")
		}())
	}

	// case not in go mod and go path
	err = g.Generate("./test.proto", projectDir, []string{src})
	assert.Nil(t, err)
	_, err = execx.Run("go test "+projectName, projectDir)
	if err != nil {
		assert.True(t, func() bool {
			return strings.Contains(err.Error(), "not in GOROOT") || strings.Contains(err.Error(), "cannot find package")
		}())
	}
}
