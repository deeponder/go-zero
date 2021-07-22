package template

var (
	// Imports defines a import template for model in cache case
	Imports = `import (
	"database/sql"
	"fmt"
	"strings"
	{{if .time}}"time"{{end}}

	"gitlab.deepwisdomai.com/infra/go-zero/core/stores/cache"
	"gitlab.deepwisdomai.com/infra/go-zero/core/stores/sqlc"
	"gitlab.deepwisdomai.com/infra/go-zero/core/stores/sqlx"
	"gitlab.deepwisdomai.com/infra/go-zero/core/stringx"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/model/sql/builderx"
)
`
	// ImportsNoCache defines a import template for model in normal case
	ImportsNoCache = `import (
	"database/sql"
	"fmt"
	"strings"
	{{if .time}}"time"{{end}}

	"gitlab.deepwisdomai.com/infra/go-zero/core/stores/sqlc"
	"gitlab.deepwisdomai.com/infra/go-zero/core/stores/sqlx"
	"gitlab.deepwisdomai.com/infra/go-zero/core/stringx"
	"gitlab.deepwisdomai.com/infra/go-zero/tools/goctl/model/sql/builderx"
)
`
)
