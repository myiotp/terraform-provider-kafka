package kafka

import "testing"

func TestMapEq(t *testing.T) {
	a := map[string]string{
		"a": "foo",
	}

	b := map[string]string{
		"a": "foo",
	}

	err := MapEq(a, b)
	if err != nil {
		t.Fatalf("%s", err)
	}
}
