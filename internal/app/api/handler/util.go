package handler

func convertArrayToString(emails []string) string {
	result := emails[0]
	for i := 1; i < len(emails); i++ {
		result += ";" + emails[i]
	}
	return result
}
