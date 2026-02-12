package version

import (
	_ "embed"
	"runtime/debug"
	"strings"
)

var (
	//go:embed version.txt
	version string
	commit  string
	dirty   bool
	// Set at build time via -ldflags for dev builds
	branch    string
	buildTime string
	buildUser string
)

func init() {
	version = strings.TrimSpace(version)
	// Attempt to get build info from the Go runtime. We only use this if not
	// built from a tagged version.
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version == "(devel)" {
		commit = getCommit(info)
		dirty = getDirty(info)
	}
}

func getDirty(info *debug.BuildInfo) bool {
	for _, setting := range info.Settings {
		if setting.Key == "vcs.modified" {
			return setting.Value == "true"
		}
	}
	return false
}

func getCommit(info *debug.BuildInfo) string {
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value[:7]
		}
	}
	return ""
}

// GetVersion returns the version of Task. By default, this is retrieved from
// the embedded version.txt file which is kept up-to-date by our release script.
// However, it can also be overridden at build time using:
// -ldflags="-X 'github.com/go-task/task/v3/internal/version.version=vX.X.X'".
func GetVersion() string {
	return version
}

// GetVersionWithBuildInfo is the same as [GetVersion], but it also includes
// the commit hash and dirty status if available. This will only work when built
// within inside of a Git checkout.
func GetVersionWithBuildInfo() string {
	var buildMetadata []string
	if commit != "" {
		buildMetadata = append(buildMetadata, commit)
	}
	if dirty {
		buildMetadata = append(buildMetadata, "dirty")
	}
	if len(buildMetadata) > 0 {
		return version + "+" + strings.Join(buildMetadata, ".")
	}
	return version
}

// GetDetailedVersionInfo returns a multi-line version string with git branch,
// commit, build time, and user. Useful for dev builds to identify the exact
// source of the binary.
func GetDetailedVersionInfo() string {
	var sb strings.Builder
	sb.WriteString("Task version: ")
	sb.WriteString(GetVersionWithBuildInfo())
	if branch != "" {
		sb.WriteString("\nBranch:       ")
		sb.WriteString(branch)
	}
	if commit != "" {
		sb.WriteString("\nCommit:       ")
		sb.WriteString(commit)
	}
	if dirty {
		sb.WriteString(" (dirty)")
	}
	if buildTime != "" {
		sb.WriteString("\nBuilt:        ")
		sb.WriteString(buildTime)
	}
	if buildUser != "" {
		sb.WriteString("\nBuilt by:     ")
		sb.WriteString(buildUser)
	}
	return sb.String()
}
