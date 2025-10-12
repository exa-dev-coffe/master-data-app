package utils

import "strconv"

func StringToInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0 // atau handle error sesuai kebutuhan
	}
	return n
}

func IntToString(i int) string {
	return strconv.Itoa(i)
}
