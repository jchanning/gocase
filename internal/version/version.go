package version

// Version is the application version, injected at build time via:
//
//	go build -ldflags="-X github.com/jchanning/gocase/internal/version.Version=1.0.0"
var Version = "dev"
