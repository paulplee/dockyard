// Package dockyard is a thin root package whose only job is to embed the
// templates/ and shared/ directories so they travel with the compiled binary.
// No runtime logic lives here — subpackages consume Assets directly.
package dockyard

import "embed"

// Assets is the embedded filesystem containing every template's build context
// plus the shared/ tree (entrypoint.sh etc.). The path prefixes inside Assets
// match their on-disk layout: "templates/<name>/..." and "shared/...".
//
//go:embed all:templates all:shared
var Assets embed.FS
