package kafka

import (
	"fmt"
)

// MapEq compares two maps, and checks that the keys and values are the same
func MapEq(result, expected map[string]string) error {
	if len(result) != len(expected) {
		return fmt.Errorf("%v != %v", result, expected)
	}

	for expectedK, expectedV := range expected {
		if resultV, ok := result[expectedK]; ok {
			if resultV != expectedV {
				return fmt.Errorf("result[%s]: %s != expected[%s]: %s", expectedK, resultV, expectedK, expectedV)
			}

		} else {
			return fmt.Errorf("result[%s] should exist", expectedK)
		}
	}
	return nil
}

func MapToPtrMap(m map[string]string) map[string]*string {
	m2 := make(map[string]*string)
	for k, v := range m {
		m2[k] = &v
	}

	return m2
}
