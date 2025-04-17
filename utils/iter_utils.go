package utils

func Map[T any, U any](list []T, mapFn func(value T) U) []U {
	mappedValues := make([]U, len(list))
	for i, value := range list {
		mappedValue := mapFn(value)
		mappedValues[i] = mappedValue
	}
	return mappedValues
}

func Filter[T any](list []T, filterFn func(value T) bool) []T {
	filteredValues := []T{}
	for _, value := range list {
		includeValue := filterFn(value)
		if includeValue {
			filteredValues = append(filteredValues, value)
		}
	}
	return filteredValues
}
