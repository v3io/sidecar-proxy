package common

func StringInSlice(s string, slice []string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}
