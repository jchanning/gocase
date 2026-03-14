package version

// Version is the application version, injected at build time via:
//
//	go build -ldflags="-X my-app/internal/version.Version=1.0.0"
var Version = "dev"
