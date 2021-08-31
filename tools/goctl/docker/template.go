package docker

import (
	"github.com/urfave/cli"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util"
)

const (
	category              = "docker"
	dockerTemplateFile    = "docker.tpl"
	ciTemplateFile        = "ci.tpl"
	gitIgnoreTemplateFile = "gitignore.tpl"
	dockerTemplate        = `FROM ccr.deepwisdomai.com/infra/go-base-image/go-builder:1.0-alpine AS builder

LABEL stage=gobuilder

# go的环境变量
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct
ENV GOPRIVATE *.deepwisdomai.com
ENV GO111MODULE on

WORKDIR /build/dw

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
{{if .Argument}}COPY {{.GoRelPath}}/etc /app/etc
{{end}}RUN go build -ldflags="-s -w" -o /app/{{.ExeFile}} {{.GoRelPath}}/{{.GoFile}}


FROM ccr.deepwisdomai.com/pub/ci/alpine
RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/{{.ExeFile}} /app/{{.ExeFile}}{{if .Argument}}
COPY --from=builder /app/etc /app/etc{{end}}
{{if .HasPort}}
EXPOSE {{.Port}}
{{end}}
CMD ["./{{.ExeFile}}"]
`
)

// Clean deletes all templates files
func Clean() error {
	return util.Clean(category)
}

// GenTemplates creates docker template files
func GenTemplates(_ *cli.Context) error {
	return initTemplate()
}

// Category returns the const string of docker category
func Category() string {
	return category
}

// RevertTemplate recovers the deleted template files
func RevertTemplate(name string) error {
	return util.CreateTemplate(category, name, dockerTemplate)
}

// Update deletes and creates new template files
func Update() error {
	err := Clean()
	if err != nil {
		return err
	}

	return initTemplate()
}

func initTemplate() error {
	return util.InitTemplates(category, map[string]string{
		dockerTemplateFile: dockerTemplate,
	})
}
