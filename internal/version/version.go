package version

// Version information set at build time via -ldflags
var (
	// Version is the current version of wui
	Version = "dev"

	// Commit is the git commit hash at build time
	Commit = "unknown"

	// BuildDate is the date when the binary was built
	BuildDate = "unknown"
)

// GetVersion returns the full version string
func GetVersion() string {
	if Commit != "unknown" && BuildDate != "unknown" {
		return Version + " (" + Commit[:7] + ", " + BuildDate + ")"
	}
	return Version
}
