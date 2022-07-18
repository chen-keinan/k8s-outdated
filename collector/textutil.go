package collector

import "strings"

func FindRemovedDeprecatedVersion(lower string, verb string) string {
	dIndex := strings.Index(lower, verb)
	ndes := lower[dIndex+len(verb):]
	sndes := strings.Split(strings.TrimPrefix(ndes, " "), " ")
	rem := strings.TrimSuffix(strings.TrimSuffix(sndes[0], ","), ".")
	return rem
}
