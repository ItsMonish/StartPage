package embeddings

import "embed"

//go:embed web/*
var StaticAssets embed.FS

//go:embed template/startpage.html
var TemplateHTML string
