package buildinfo

// AppVersion is the application version. Injected at build time via
// -ldflags "-X nodus/pkg/buildinfo.AppVersion=..." and falls back to the
// value below when running from source.
var AppVersion = "0.0.4"

// BuildTime is the build timestamp. Injected at build time via ldflags.
var BuildTime = "unknown"
