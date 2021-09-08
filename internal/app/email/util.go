package email

import "strings"

func ConvertArrayToString(emails []string) string {
	return strings.Join(emails, ";")
}
