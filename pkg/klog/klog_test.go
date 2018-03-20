/*
Copyright 2018 ZTE Corporation. All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package klog

import (
	"errors"
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"sync"
	"testing"
)

func B() {
	fmt.Println("B", goId())
}
func A() {
	fmt.Println("A", goId())
	B()
}
func Testklog() int {

	fmt.Println("test", goId())
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(i, goId())
		}()
	}
	A()
	wg.Wait()

	err := errors.New("this is err!")
	Errorf("loginfo:%v%s", err, "this is problem")
	//Errorf("loginfo:", err, "this is problem")
	Error("loginfo:", err, "this is problem")

	return 0
}

func Test_klog_Normal(t *testing.T) {

	ret := Testklog()

	convey.Convey("Subject:Test_klog_Normal\n", t, func() {
		convey.Convey("Testklog Exec", func() {
			convey.So(ret, convey.ShouldEqual, 0)
		})
	})
}
