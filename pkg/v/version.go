package v

import "fmt"

var (
	gitCommit  = ""
	buildTime  = ""
	goVersion  = ""
	appVersion = "1.0.0"
)

// Version returns the full version information for the application.
func Version() string {
	return fmt.Sprintf("version: %s\n", appVersion) +
		fmt.Sprintf("build  : %s\n", buildTime) +
		fmt.Sprintf("git    : %s\n", gitCommit) +
		fmt.Sprintf("go     : %s", goVersion)
}
