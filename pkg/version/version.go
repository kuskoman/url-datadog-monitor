package version

// These variables are set during build time via -ldflags
var (
	// Version is the version of the application
	Version = "dev"
	
	// GitCommit is the git commit hash
	GitCommit = "unknown"
	
	// BuildDate is the date the application was built
	BuildDate = "unknown"
)

// GetVersionInfo returns the version information as a map
func GetVersionInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"gitCommit":  GitCommit,
		"buildDate":  BuildDate,
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	return Version + " (commit: " + GitCommit + ", build date: " + BuildDate + ")"
}