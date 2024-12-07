package utils

func FindIndex(rawData []byte, startIndex int, target byte) int {
	for i := startIndex; i < len(rawData); i++ {
		if rawData[i] == target {
			return i
		}
	}
	return -1
}

func Prepend(rawData []byte, values ...byte) []byte {

	var result []byte
	if len(values) == 1 {
		result = []byte{values[0]}
	} else {
		result = values
	}

	return append(result, rawData...)
}

func Append(rawData []byte, values ...byte) []byte {

	var result []byte
	if len(values) == 1 {
		result = []byte{values[0]}
	} else {
		result = values
	}

	return append(rawData, result...)
}

func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func RemoveFromSlice(slice []string, str string) []string {
	var result []string
	for _, s := range slice {
		if s != str {
			result = append(result, s)
		}
	}
	return result
}
