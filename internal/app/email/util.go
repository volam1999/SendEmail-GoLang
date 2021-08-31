package email

func ConvertArrayToString(emails []string) string {
	if len(emails) == 0 {
		return ""
	}
	result := emails[0]
	for i := 1; i < len(emails); i++ {
		result += ";" + emails[i]
	}
	return result
}
