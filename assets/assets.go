package assets

import "embed"

//go:embed web/*
var WebFS embed.FS

//go:embed migrations/001_initial.sql
var MigrationSQL []byte
