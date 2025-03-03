package command

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/urfave/cli"
	"gitlab.deepwisdomai.com/infra/go-zero/core/logx"
	"gitlab.deepwisdomai.com/infra/go-zero/core/stores/sqlx"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/config"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/model/sql/gen"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/model/sql/model"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/model/sql/util"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/util/console"
)

const (
	flagSrc   = "src"
	flagDir   = "dir"
	flagCache = "cache"
	flagIdea  = "idea"
	flagURL   = "url"
	flagTable = "table"
	flagStyle = "style"
)

var errNotMatched = errors.New("sql not matched")

// MysqlDDL generates model code from ddl
func MysqlDDL(ctx *cli.Context) error {
	src := ctx.String(flagSrc)
	dir := ctx.String(flagDir)
	cache := ctx.Bool(flagCache)
	idea := ctx.Bool(flagIdea)
	style := ctx.String(flagStyle)
	cfg, err := config.NewConfig(style)
	if err != nil {
		return err
	}

	return fromDDl(src, dir, cfg, cache, idea)
}

// MyDataSource generates model code from datasource
func MyDataSource(ctx *cli.Context) error {
	url := strings.TrimSpace(ctx.String(flagURL))
	dir := strings.TrimSpace(ctx.String(flagDir))
	cache := ctx.Bool(flagCache)
	idea := ctx.Bool(flagIdea)
	style := ctx.String(flagStyle)
	pattern := strings.TrimSpace(ctx.String(flagTable))
	cfg, err := config.NewConfig(style)
	if err != nil {
		return err
	}

	return fromDataSource(url, pattern, dir, cfg, cache, idea)
}

func fromDDl(src, dir string, cfg *config.Config, cache, idea bool) error {
	log := console.NewConsole(idea)
	src = strings.TrimSpace(src)
	if len(src) == 0 {
		return errors.New("expected path or path globbing patterns, but nothing found")
	}

	files, err := util.MatchFiles(src)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errNotMatched
	}

	var source []string
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		source = append(source, string(data))
	}

	generator, err := gen.NewDefaultGenerator(dir, cfg, gen.WithConsoleOption(log))
	if err != nil {
		return err
	}

	return generator.StartFromDDL(strings.Join(source, "\n"), cache)
}

func fromDataSource(url, pattern, dir string, cfg *config.Config, cache, idea bool) error {
	log := console.NewConsole(idea)
	if len(url) == 0 {
		log.Error("%v", "expected data source of mysql, but nothing found")
		return nil
	}

	if len(pattern) == 0 {
		log.Error("%v", "expected table or table globbing patterns, but nothing found")
		return nil
	}

	dsn, err := mysql.ParseDSN(url)
	if err != nil {
		return err
	}

	logx.Disable()
	databaseSource := strings.TrimSuffix(url, "/"+dsn.DBName) + "/information_schema"
	db := sqlx.NewMysql(databaseSource)
	im := model.NewInformationSchemaModel(db)

	tables, err := im.GetAllTables(dsn.DBName)
	if err != nil {
		return err
	}

	matchTables := make(map[string]*model.Table)
	for _, item := range tables {
		match, err := filepath.Match(pattern, item)
		if err != nil {
			return err
		}

		if !match {
			continue
		}

		columnData, err := im.FindColumns(dsn.DBName, item)
		if err != nil {
			return err
		}

		table, err := columnData.Convert()
		if err != nil {
			return err
		}

		matchTables[item] = table
	}

	if len(matchTables) == 0 {
		return errors.New("no tables matched")
	}

	generator, err := gen.NewDefaultGenerator(dir, cfg, gen.WithConsoleOption(log))
	if err != nil {
		return err
	}

	return generator.StartFromInformationSchema(matchTables, cache)
}
