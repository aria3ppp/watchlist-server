package token

import "strings"

func ExtractTokenFromAuth(auth string) string {
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}
