package version

import (
	"fmt"
	"runtime"
)

// tag represents the Git tag associated with the build.
// If the build is not tagged, the value will be set to "unknown".
// Set this value while compiling the binary by passing the -ldflags "-X version.tag=<tag>" flag to the go build command.
var tag = "unknown"

// revision represents the Git commit associated with the build.
// If the build is not tagged, the value will be set to "unknown".
// Set this value while compiling the binary by passing the -ldflags "-X version.revision=<commit>" flag to the go build command.
var revision = "unknown"

// buildDate represents the date and time the build was created.
// If the build is not tagged, the value will be set to "unknown".
// Set this value while compiling the binary by passing the -ldflags "-X version.buildDate=<date>" flag to the go build command.
var buildDate = "unknown"

// buildHost represents the host machine on which the build was created.
// If the build is not tagged, the value will be set to "unknown".
// Set this value while compiling the binary by passing the -ldflags "-X version.buildHost=<hostname>" flag to the go build command.
var buildHost = "unknown"

// buildUser represents the user who created the build.
// If the build is not tagged, the value will be set to "unknown".
// Set this value while compiling the binary by passing the -ldflags "-X version.buildUser=<username>" flag to the go build command.
var buildUser = "unknown"

// isTaint represents whether the git has uncommitted changes or not.
// If the build is tainted, the value will be set to "1".
// If the build is not tainted, the value will be set to "0".
// Set this value while compiling the binary by passing the -ldflags "-X version.isTaint=<taint>" flag to the go build command.
var isTaint = "-1"

// BuildInfo is a struct that represents the build information of the application.
type BuildInfo struct {
	Tag       string `json:"tag,omitempty"`           // Git tag associated with the build
	Revision  string `json:"revision,omitempty"`      // Git commit associated with the build
	BuildDate string `json:"build_date,omitempty"`    // Date and time the build was created
	BuildHost string `json:"build_machine,omitempty"` // Host machine on which the build was created
	BuildUser string `json:"build_user,omitempty"`    // User who created the build
	IsTaint   string `json:"is_taint,omitempty"`      // Whether the build is tainted or not
	GoArch    string `json:"go_arch,omitempty"`       // compiled binary CPU architecture
	GoOS      string `json:"go_os,omitempty"`         // compiled binary OS target
	GoVersion string `json:"go_version,omitempty"`    // Go version used to compile the binary
}

// Info returns a BuildInfo object with the current build information.
func (bi BuildInfo) Info() BuildInfo {
	return BuildInfo{
		Tag:       tag,
		Revision:  revision,
		BuildDate: buildDate,
		BuildHost: buildHost,
		BuildUser: buildUser,
		IsTaint:   isTaint,
		GoArch:    runtime.GOARCH,
		GoOS:      runtime.GOOS,
		GoVersion: runtime.Version(),
	}
}

// String returns a formatted string representation of the BuildInfo object.
func (bi BuildInfo) String() string {
	return fmt.Sprintf("Tag: %s, Revision: %s, BuildDate: %s, BuildHost: %s, BuildUser: %s, IsTaint: %s, GoArch: %s, GoOS: %s, GoVersion: %s",
		bi.Tag,
		bi.Revision,
		bi.BuildDate,
		bi.BuildHost,
		bi.BuildUser,
		bi.IsTaint,
		bi.GoArch,
		bi.GoOS,
		bi.GoVersion)
}
