package common

import (
	"fmt"
)

// FooBarInvocation just for testing
func FooBarInvocation(str string) error {
	if str == "true" {
		fmt.Println("correct! man :)")
		fmt.Println("test")
		return nil
	}
	fmt.Println("incorrect! man!!!")
	return fmt.Errorf("incorrect! man :<")
}
