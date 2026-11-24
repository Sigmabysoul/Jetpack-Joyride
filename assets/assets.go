package assets

import "embed"

//go:embed *.png *.jpg *.ttf
var FS embed.FS
