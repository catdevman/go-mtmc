package disk

import "embed"

// FS represents the embedded disk filesystem.
// The `all:` prefix includes all files in the directory,
// including those that start with a `.` or `_`.

//go:embed disk
var FS embed.FS
