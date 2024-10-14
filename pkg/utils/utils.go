package utils

import ()

func All[K comparable, V any](m map[K]V, condition func(V) bool) bool {
	for _, value := range m {
		if !condition(value) {
			return false // Return false if any value doesn't meet the condition
		}
	}
	return true // Return true if all values meet the condition
}
