package utils

func FindIndex(rawData []byte, startIndex int, target byte) int {
	for i := startIndex; i < len(rawData); i++ {
		if rawData[i] == target {
			return i
		}
	}
	return -1
}
