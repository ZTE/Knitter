package alg

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIntSliceInit(t *testing.T) {
	slice := NewSlice()
	convey.Convey("init a int slice and len should equal 0\n", t, func() {
		convey.So(len(slice), convey.ShouldEqual, 0)
	})
}

func TestIntSliceAddWithNotRepeated(t *testing.T) {
	slice := NewSlice()

	convey.Convey("add three int value with not repeted\n", t, func() {
		slice.Add(1)
		convey.So(slice[0], convey.ShouldEqual, 1)
		slice.Add(2)
		convey.So(slice[1], convey.ShouldEqual, 2)
		slice.Add(3)
		convey.So(slice[2], convey.ShouldEqual, 3)
		convey.So(len(slice), convey.ShouldEqual, 3)

	})
}

func TestIntSliceAddWithRepeated(t *testing.T) {
	slice := NewSlice()

	convey.Convey("add three int value with repeted\n", t, func() {
		err := slice.Add(1)
		convey.So(nil, convey.ShouldEqual, err)
		convey.So(slice[0], convey.ShouldEqual, 1)
		err = slice.Add(2)
		convey.So(nil, convey.ShouldEqual, err)
		convey.So(slice[1], convey.ShouldEqual, 2)
		err = slice.Add(2)
		convey.So(ErrElemExist.Error(), convey.ShouldEqual, err.Error())
		err = slice.Add(3)
		convey.So(nil, convey.ShouldEqual, err)
		convey.So(slice[2], convey.ShouldEqual, 3)
		convey.So(len(slice), convey.ShouldEqual, 3)

	})
}

func TestIntSliceRemoveWithNotRepeated(t *testing.T) {
	slice := NewSlice()
	slice.Add(1)
	slice.Add(2)
	slice.Add(3)

	convey.Convey("remove three int value with not repeated\n", t, func() {
		convey.So(len(slice), convey.ShouldEqual, 3)
		slice.Remove(2)
		convey.So(len(slice), convey.ShouldEqual, 2)
		slice.Remove(3)
		convey.So(len(slice), convey.ShouldEqual, 1)
		slice.Remove(1)
		convey.So(len(slice), convey.ShouldEqual, 0)
	})
}

func TestIntSliceRemoveWithRepeated(t *testing.T) {
	slice := NewSlice()
	slice.Add(1)
	slice.Add(2)
	slice.Add(3)

	convey.Convey("remove three int value with not repeated\n", t, func() {
		convey.So(len(slice), convey.ShouldEqual, 3)
		err := slice.Remove(2)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(slice), convey.ShouldEqual, 2)
		convey.So(slice[0], convey.ShouldEqual, 1)
		convey.So(slice[1], convey.ShouldEqual, 3)
		err = slice.Remove(2)
		convey.So(err.Error(), convey.ShouldEqual, ErrElemNtExist.Error())
		convey.So(len(slice), convey.ShouldEqual, 2)
		err = slice.Remove(3)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(slice), convey.ShouldEqual, 1)
		convey.So(slice[0], convey.ShouldEqual, 1)
		err = slice.Remove(1)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(slice), convey.ShouldEqual, 0)
	})
}

func TestStringSliceInit(t *testing.T) {
	slice := NewSlice()
	convey.Convey("init a string slice and len should equal 0\n", t, func() {
		convey.So(len(slice), convey.ShouldEqual, 0)
	})
}

func TestStringSliceAddWithNotRepeated(t *testing.T) {
	str1 := "hello"
	str2 := "golang"
	str3 := "generic"
	slice := NewSlice()

	convey.Convey("add three string value with not repeted\n", t, func() {
		slice.Add(str1)
		convey.So(slice[0], convey.ShouldEqual, str1)
		slice.Add(str2)
		convey.So(slice[1], convey.ShouldEqual, str2)
		slice.Add(str3)
		convey.So(slice[2], convey.ShouldEqual, str3)
		convey.So(len(slice), convey.ShouldEqual, 3)

	})
}

func TestStringSliceAddWithRepeated(t *testing.T) {
	str1 := "hello"
	str2 := "golang"
	str3 := "generic"
	slice := NewSlice()

	convey.Convey("add three string value with repeted\n", t, func() {
		err := slice.Add(str1)
		convey.So(nil, convey.ShouldEqual, err)
		convey.So(slice[0], convey.ShouldEqual, str1)
		err = slice.Add(str2)
		convey.So(nil, convey.ShouldEqual, err)
		convey.So(slice[1], convey.ShouldEqual, str2)
		err = slice.Add(str2)
		convey.So(ErrElemExist.Error(), convey.ShouldEqual, err.Error())
		err = slice.Add(str3)
		convey.So(nil, convey.ShouldEqual, err)
		convey.So(slice[2], convey.ShouldEqual, str3)
		convey.So(len(slice), convey.ShouldEqual, 3)

	})
}

func TestStringSliceRemoveWithNotRepeated(t *testing.T) {
	str1 := "hello"
	str2 := "golang"
	str3 := "generic"
	slice := NewSlice()
	slice.Add(str1)
	slice.Add(str2)
	slice.Add(str3)

	convey.Convey("remove three string value with not repeated\n", t, func() {
		convey.So(len(slice), convey.ShouldEqual, 3)
		slice.Remove(str2)
		convey.So(len(slice), convey.ShouldEqual, 2)
		slice.Remove(str3)
		convey.So(len(slice), convey.ShouldEqual, 1)
		slice.Remove(str1)
		convey.So(len(slice), convey.ShouldEqual, 0)
	})
}

func TestStringSliceRemoveWithRepeated(t *testing.T) {
	str1 := "hello"
	str2 := "golang"
	str3 := "generic"
	slice := NewSlice()
	slice.Add(str1)
	slice.Add(str2)
	slice.Add(str3)

	convey.Convey("remove three string value with not repeated\n", t, func() {
		convey.So(len(slice), convey.ShouldEqual, 3)
		err := slice.Remove(str2)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(slice), convey.ShouldEqual, 2)
		convey.So(slice[0], convey.ShouldEqual, str1)
		convey.So(slice[1], convey.ShouldEqual, str3)
		err = slice.Remove(str2)
		convey.So(err.Error(), convey.ShouldEqual, ErrElemNtExist.Error())
		convey.So(len(slice), convey.ShouldEqual, 2)
		convey.So(slice[0], convey.ShouldEqual, str1)
		convey.So(slice[1], convey.ShouldEqual, str3)
		err = slice.Remove(str3)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(slice), convey.ShouldEqual, 1)
		convey.So(slice[0], convey.ShouldEqual, str1)
		err = slice.Remove(str1)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(slice), convey.ShouldEqual, 0)
	})
}

type Student struct {
	id   string
	name string
}

func (this Student) IsEqual(obj interface{}) bool {
	if student, ok := obj.(Student); ok {
		return this.id == student.GetID()
	}
	panic("unexpected type")
}

func (this Student) GetID() string {
	return this.id
}

func TestStructSliceAddWithNotRepeated(t *testing.T) {
	obj1 := Student{"1001", "xiao ming"}
	obj2 := Student{"1002", "xiao lei"}
	obj3 := Student{"1003", "xiao fang"}
	slice := NewSlice()

	convey.Convey("add three struct value with not repeted\n", t, func() {
		slice.Add(obj1)
		convey.So(slice[0].(Student).GetID(), convey.ShouldEqual, obj1.GetID())
		slice.Add(obj2)
		convey.So(slice[1].(Student).GetID(), convey.ShouldEqual, obj2.GetID())
		slice.Add(obj3)
		convey.So(slice[2].(Student).GetID(), convey.ShouldEqual, obj3.GetID())
		convey.So(len(slice), convey.ShouldEqual, 3)

	})
}

func TestStructSliceAddWithRepeated(t *testing.T) {
	obj1 := Student{"1001", "xiao ming"}
	obj2 := Student{"1002", "xiao lei"}
	obj3 := Student{"1003", "xiao fang"}
	slice := NewSlice()

	convey.Convey("add three struct value with repeted\n", t, func() {
		err := slice.Add(obj1)
		convey.So(nil, convey.ShouldEqual, err)
		convey.So(slice[0].(Student).GetID(), convey.ShouldEqual, obj1.GetID())
		err = slice.Add(obj2)
		convey.So(nil, convey.ShouldEqual, err)
		convey.So(slice[1].(Student).GetID(), convey.ShouldEqual, obj2.GetID())
		err = slice.Add(obj2)
		convey.So(ErrElemExist.Error(), convey.ShouldEqual, err.Error())
		err = slice.Add(obj3)
		convey.So(nil, convey.ShouldEqual, err)
		convey.So(slice[2].(Student).GetID(), convey.ShouldEqual, obj3.GetID())
		convey.So(len(slice), convey.ShouldEqual, 3)

	})
}

func TestStructSliceRemoveWithNotRepeated(t *testing.T) {
	obj1 := Student{"1001", "xiao ming"}
	obj2 := Student{"1002", "xiao lei"}
	obj3 := Student{"1003", "xiao fang"}
	slice := NewSlice()
	slice.Add(obj1)
	slice.Add(obj2)
	slice.Add(obj3)

	convey.Convey("remove three struct value with not repeated\n", t, func() {
		convey.So(len(slice), convey.ShouldEqual, 3)
		slice.Remove(obj2)
		convey.So(len(slice), convey.ShouldEqual, 2)
		slice.Remove(obj3)
		convey.So(len(slice), convey.ShouldEqual, 1)
		slice.Remove(obj1)
		convey.So(len(slice), convey.ShouldEqual, 0)
	})
}

func TestStructSliceRemoveWithRepeated(t *testing.T) {
	obj1 := Student{"1001", "xiao ming"}
	obj2 := Student{"1002", "xiao lei"}
	obj3 := Student{"1003", "xiao fang"}
	slice := NewSlice()
	slice.Add(obj1)
	slice.Add(obj2)
	slice.Add(obj3)

	convey.Convey("remove three struct value with not repeated\n", t, func() {
		convey.So(len(slice), convey.ShouldEqual, 3)
		err := slice.Remove(obj2)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(slice), convey.ShouldEqual, 2)
		convey.So(slice[0].(Student).GetID(), convey.ShouldEqual, obj1.GetID())
		convey.So(slice[1].(Student).GetID(), convey.ShouldEqual, obj3.GetID())
		err = slice.Remove(obj2)
		convey.So(err.Error(), convey.ShouldEqual, ErrElemNtExist.Error())
		convey.So(len(slice), convey.ShouldEqual, 2)
		convey.So(slice[0].(Student).GetID(), convey.ShouldEqual, obj1.GetID())
		convey.So(slice[1].(Student).GetID(), convey.ShouldEqual, obj3.GetID())
		err = slice.Remove(obj3)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(slice), convey.ShouldEqual, 1)
		convey.So(slice[0].(Student).GetID(), convey.ShouldEqual, obj1.GetID())
		err = slice.Remove(obj1)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(slice), convey.ShouldEqual, 0)
	})
}
