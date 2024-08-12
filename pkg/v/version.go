package v

import "fmt"

var (
	GitCommit  = ""
	BuildTime  = ""
	BuildHost  = ""
	BuildIP  = ""
	GoVersion  = ""
	AppVersion = "1.0.0"
)

// Version returns the full version information for the application.
func Version() string {
	return "" +
		fmt.Sprintf("version    : %s\n", AppVersion) +
		fmt.Sprintf("build at   : %s\n", BuildTime) +
		fmt.Sprintf("build host : %s\n", BuildHost) +
		fmt.Sprintf("build ip   : %s\n", BuildIP) +
		fmt.Sprintf("git        : %s\n", GitCommit) +
		fmt.Sprintf("go         : %s", GoVersion)
}
