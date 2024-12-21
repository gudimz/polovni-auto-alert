package utils

// RemoveDuplicates removes duplicates from a slice of any type.
func RemoveDuplicates[T comparable](slice []T) []T {
	seen := make(map[T]struct{})

	var result []T

	for _, item := range slice {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}

			result = append(result, item)
		}
	}

	return result
}
