package version

import "runtime/debug"

var readBuildInfo = debug.ReadBuildInfo

func Version() string {
	if version, ok := readBuildInfo(); ok {
		return version.Main.Version
	}
	return "unknown"
}
