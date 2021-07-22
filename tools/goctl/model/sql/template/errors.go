package template

// Error defines an error template
var Error = `package {{.pkg}}

import "gitlab.deepwisdomai.com/infra/go-zero/core/stores/sqlx"

var ErrNotFound = sqlx.ErrNotFound
`
