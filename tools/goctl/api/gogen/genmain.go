package gogen

import (
	"fmt"
	"strings"

	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/api/spec"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/config"
	ctlutil "gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util/format"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/vars"
)

const mainTemplate = `package main

import (
	"flag"
	"fmt"
	"os"
	"log"
	
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"gitlab.deepwisdomai.com/infra/go-zero/core/logx"
	
	{{.importPackages}}
)

var configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "the config file")

func main() {
	flag.Parse()

	var (
		c config.Config
		dc config.DynamicConfig
	)
	
	ctx := svc.NewServiceContext(c, dc)

	if env := os.Getenv("ENV"); env != "" {
		*configFile = "etc/{{.serviceName}}-" + env + ".yaml"
	}
	conf.MustLoad(*configFile, &c)
	ctx.Config = c

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

		if err = conf.LoadConfigFromYamlBytes([]byte(content), &dc); err != nil {
			log.Fatal(err)
		}
		ctx.DynamicConfig = dc
		
		// listen nacos conf change
		err = client.ListenConfig(vo.ConfigParam{
			DataId: "{{.serviceName}}.yaml",
			Group:  "DEFAULT_GROUP",
			OnChange: func(namespace, group, dataId, data string) {
				var newConf config.DynamicConfig
				if err = conf.LoadConfigFromYamlBytes([]byte(data), &newConf); err != nil {
					logx.Errorf("update dynamic conf err:%s", err.Error())
					return
				}

				ctx.DynamicConfig = newConf
			},
		})
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
`

func genMain(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	name := strings.ToLower(api.Service.Name)
	if strings.HasSuffix(name, "-api") {
		name = strings.ReplaceAll(name, "-api", "")
	}
	filename, err := format.FileNamingFormat(cfg.NamingFormat, name)
	if err != nil {
		return err
	}

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          "",
		filename:        filename + ".go",
		templateName:    "mainTemplate",
		category:        category,
		templateFile:    mainTemplateFile,
		builtinTemplate: mainTemplate,
		data: map[string]string{
			"importPackages": genMainImports(rootPkg),
			"serviceName":    api.Service.Name,
		},
	})
}

func genMainImports(parentPkg string) string {
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s\"", ctlutil.JoinPackages(parentPkg, configDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"", ctlutil.JoinPackages(parentPkg, handlerDir)))
	imports = append(imports, fmt.Sprintf("\"%s\"\n", ctlutil.JoinPackages(parentPkg, contextDir)))
	imports = append(imports, fmt.Sprintf("\"%s/core/conf\"", vars.ProjectOpenSourceURL))
	imports = append(imports, fmt.Sprintf("\"%s/rest\"", vars.ProjectOpenSourceURL))
	return strings.Join(imports, "\n\t")
}
