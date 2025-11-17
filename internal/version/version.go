package version

const defaultVersion = "dev"

// Version is the ComposePack CLI version. It defaults to defaultVersion but can be overridden via -ldflags.
var Version = defaultVersion
