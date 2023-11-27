package helpers

import (
	"reflect"
	"strings"
)

func AreTagsEqual(tags1, tags2 []string) bool {
	if len(tags1) != len(tags2) {
		return false
	}
	tagMap := make(map[string]bool)
	for _, tag := range tags1 {
		tagMap[tag] = true
	}
	for _, tag := range tags2 {
		if !tagMap[tag] {
			return false
		}
	}
	return true
}

func AreSlicesEqual(slice1, slice2 interface{}) bool {
	s1 := reflect.ValueOf(slice1)
	s2 := reflect.ValueOf(slice2)

	// Check if both inputs are slices
	if s1.Kind() != reflect.Slice || s2.Kind() != reflect.Slice {
		return false
	}

	// Check slice lengths
	if s1.Len() != s2.Len() {
		return false
	}

	// Create a map to track elements in slice1
	elementsMap := make(map[interface{}]bool)
	for i := 0; i < s1.Len(); i++ {
		elementsMap[s1.Index(i).Interface()] = true
	}

	// Check if elements in slice2 are in slice1
	for i := 0; i < s2.Len(); i++ {
		if !elementsMap[s2.Index(i).Interface()] {
			return false
		}
	}

	return true
}

func IsResourceNotFoundErr(err error, resourceName string) bool {
	notFoundStr := resourceName + " not found"
	return strings.Contains(err.Error(), notFoundStr) || strings.Contains(err.Error(), "404")
}
