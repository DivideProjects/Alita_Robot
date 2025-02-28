package config

import (
	"strconv"
	"strings"
)

// typeConvertor is a struct that will convert a string to a specific type
type typeConvertor struct {
	str string
}

// StringArray will return a string array from a comma separated string
func (t typeConvertor) StringArray() []string {
	allUpdates := strings.Split(t.str, ",")
	for i, j := range allUpdates {
		allUpdates[i] = strings.TrimSpace(j) // this will trim the whitespace
	}
	return allUpdates
}

// Int Array will return an int array from a comma separated string
func (t typeConvertor) Int() int {
	val, _ := strconv.Atoi(t.str)
	return val
}

// Int64 Array will return an int64 array from a comma separated string
func (t typeConvertor) Int64() int64 {
	val, _ := strconv.ParseInt(t.str, 10, 64)
	return val
}

// Bool will return a bool from a string
func (t typeConvertor) Bool() bool {
	return t.str == "yes" || t.str == "true"
}
