package release

/*
 **************************************
 *** THIS FILE IS AUTO GENERATED !! ***
 **************************************
 */

type EmptyStruct struct{}

var (
	// DebugInit set init stage as debug log level in waiting config is loaded, used with -X release.DebugInit=1
	DebugInit = "0"

	// Release the git tag of the current build, used with -X release.Release=$(git describe --abbrev=0 --tags HEAD || git symbolic-ref -q --short HEAD)
	Release = "v0.0.0-rc1"

	// Build the git commit of the current build, used with -X release.Build=$(git rev-parse --short HEAD)
	Build = "a1b2c3d4"

	// Date the current datetime RFC like for the build, used with -X release.Date=$(date +%FT%T%z)
	Date = "1970-01-01T01:01:01+00:00"

	// Package is the current package name of the build directory, used with -X release.Package=$(basename $(pwd))
	Package = "auditor"

	// Description of the current package name of the build directory, used with -X release.Description=...
	Description = ""

	// Author the name of the author for the current package, used with -X release.Author=...
	Author = "github.com/nabbar/auditor"

	// Prefix the package prefix could be used example for env var, used with -X config.Prefix=...
	Prefix = "AUD"
)
