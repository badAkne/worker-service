package constant

var (
	AppName    = "worker-service"
	AppVersion = "0.1.0"
	GitCommit  = "unknown"
	BuildTime  = "unknown"
)

func GetFullVersion() string {
	return AppVersion + "()" + GitCommit + ")"
}
