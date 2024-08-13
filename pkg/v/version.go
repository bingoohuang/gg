package v

import "fmt"

var (
	GitCommit = ""
	BuildTime = ""
	BuildHost = ""
	BuildIP   = ""
	// BuildCI 构建的 CI 信息，e.g. Javis_V1.0.0_BuiltID_20240813.131
	BuildCI    = ""
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
		fmt.Sprintf("build ci   : %s\n", BuildCI) +
		fmt.Sprintf("git        : %s\n", GitCommit) +
		fmt.Sprintf("go         : %s", GoVersion)
}
