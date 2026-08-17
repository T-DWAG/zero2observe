package storage

import "errors"

// ErrNotFound 查询不到记录时返回（Memory / Postgres 统一）。
var ErrNotFound = errors.New("storage: not found")
