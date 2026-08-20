// Package skills provides the AI agent skill bundled with ggg.
package skills

import "embed"

// bundled contains the skill exactly as it is installed for the user.
//
//go:embed ggg
var bundled embed.FS
