package postgrest_go

import (
	"fmt"
	"strings"
)

const reservedChars = ",.:()"

func SanitizeParam(param string) string {
	if strings.ContainsAny(param, reservedChars) {
		return fmt.Sprintf("\"%s\"", param)
	}
	return param
}

func SanitizePatternParam(pattern string) string {
	return SanitizeParam(strings.ReplaceAll(pattern, "%", "*"))
}
