// Package version holds build-time version and release metadata.
package version

// Set at build time via ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)
