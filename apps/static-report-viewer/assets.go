package viewer

import _ "embed"

//go:embed index.html
var IndexHTML []byte

//go:embed app.js
var AppJS []byte

//go:embed styles.css
var StylesCSS []byte
