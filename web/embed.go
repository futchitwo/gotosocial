package web

import "embed"

//go:embed assets/*
//go:embed template/*
var WebFS embed.FS
