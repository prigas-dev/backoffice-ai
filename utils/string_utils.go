package utils

func IsEmpty(s string) bool {
	return len(s) == 0
}

func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}
