package web

import "embed"

//go:embed templates/*.html templates/components/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS
