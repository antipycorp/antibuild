// +build none

package site

type (
	Data  map[string]interface{}
	Value struct {
		s string
	}

	Object struct {
		value
		children
	}
)
