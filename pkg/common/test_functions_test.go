package common

import (
	"testing"
)

func TestFooBarInvocation(t *testing.T) {
	err := FooBarInvocation("true")
	if err != nil {
		t.Errorf("test error (%q)\n", err)
	}
}
