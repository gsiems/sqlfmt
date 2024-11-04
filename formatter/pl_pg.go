package formatter

import "strings"

func isPgBodyBoundary(s string) bool {
	if !strings.HasPrefix(s, "$") {
		return false
	}
	if !strings.HasSuffix(s, "$") {
		return false
	}
	if len(s) < 2 {
		return false
	}
	return true
}
