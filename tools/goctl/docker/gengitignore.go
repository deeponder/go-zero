package docker

import (
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util"
)

const (
	gitIgnoreTemplate = `### Go template
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out
### Example user template template
### Example user template

# IntelliJ project files
.idea

# VsCode project files
.vscode/*

*.log
*.zip

log/*
*.swp
*.swo
`
)

func generateGitignore() error {

	text, err := util.LoadTemplate(category, gitIgnoreTemplateFile, gitIgnoreTemplate)
	if err != nil {
		return err
	}

	return util.With("ci").Parse(text).SaveTo(nil, ".gitignore", false)
}
