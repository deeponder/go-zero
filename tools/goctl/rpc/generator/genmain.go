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

const mainTemplate = `package main

import (
	"flag"
	"fmt"
	"os"
	"log"

	{{.imports}}

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"gitlab.deepwisdomai.com/infra/go-zero/core/logx"
	
	"gitlab.deepwisdomai.com/infra/go-zero/core/conf"
	"gitlab.deepwisdomai.com/infra/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "the config file")

func main() {
	flag.Parse()

	var (
		c config.Config
	)
	ctx := svc.NewServiceContext(c)

	if env := os.Getenv("ENV"); env != "" {
		*configFile = "etc/{{.serviceName}}-" + env + ".yaml"
	}
	conf.MustLoad(*configFile, &c)

	// use nacos
	if c.NacosConf.UseNacos {
		// server conf
		sc := []constant.ServerConfig{
			{
				IpAddr: c.NacosConf.Ip,
				Port:   c.NacosConf.Port,
			},
		}

		// client conf
		cc := constant.ClientConfig{
			NamespaceId:         os.Getenv("ENV"), //namespace id
			TimeoutMs:           c.NacosConf.TimeoutMs,
			NotLoadCacheAtStart: c.NacosConf.NotLoadCacheAtStart,
			LogDir:              c.NacosConf.LogDir,
			CacheDir:            c.NacosConf.CacheDir,
			RotateTime:          c.NacosConf.RotateTime,
			MaxAge:              c.NacosConf.MaxAge,
			LogLevel:            c.NacosConf.LogLevel,
		}

		// init client
		client, err := clients.NewConfigClient(
			vo.NacosClientParam{
				ClientConfig:  &cc,
				ServerConfigs: sc,
			},
		)

		if err != nil {
			log.Fatal(err)
		}


		// get config
		content, err := client.GetConfig(vo.ConfigParam{
			DataId: "{{.serviceName}}.yaml",  //配置文件名
			Group:  "DEFAULT_GROUP",  //默认group
		})

		if err != nil {
			log.Fatal(err)
		}

		if err = conf.LoadConfigFromYamlBytes([]byte(content), &c); err != nil {
			log.Fatal(err)
		}

		// listen nacos conf change
		err = client.ListenConfig(vo.ConfigParam{
			DataId: "{{.serviceName}}.yaml",
			Group:  "DEFAULT_GROUP",
			OnChange: func(namespace, group, dataId, data string) {
				if err = conf.LoadConfigFromYamlBytes([]byte(data), &c); err != nil {
					logx.Errorf("update dynamic conf err:%s", err.Error())
					return
				}

				ctx.Config = c
			},
		})
	}

	ctx.Config = c

	srv := server.New{{.serviceNew}}Server(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		{{.pkg}}.Register{{.service}}Server(grpcServer, srv)
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
`

// GenMain generates the main file of the rpc service, which is an rpc service program call entry
func (g *DefaultGenerator) GenMain(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	mainFilename, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	fileName := filepath.Join(ctx.GetMain().Filename, fmt.Sprintf("%v.go", mainFilename))
	imports := make([]string, 0)
	pbImport := fmt.Sprintf(`"%v"`, ctx.GetPb().Package)
	svcImport := fmt.Sprintf(`"%v"`, ctx.GetSvc().Package)
	remoteImport := fmt.Sprintf(`"%v"`, ctx.GetServer().Package)
	configImport := fmt.Sprintf(`"%v"`, ctx.GetConfig().Package)
	imports = append(imports, configImport, pbImport, remoteImport, svcImport)
	text, err := util.LoadTemplate(category, mainTemplateFile, mainTemplate)
	if err != nil {
		return err
	}

	etcFileName, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	return util.With("main").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"serviceName": etcFileName,
		"imports":     strings.Join(imports, util.NL),
		"pkg":         proto.PbPackage,
		"serviceNew":  stringx.From(proto.Service.Name).ToCamel(),
		"service":     parser.CamelCase(proto.Service.Name),
	}, fileName, false)
}
