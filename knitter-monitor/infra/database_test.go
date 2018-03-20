package infra

import (
	"github.com/ZTE/Knitter/knitter-agent/tests/mock/db-mock"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetDatabaseSetDataBase(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(controller)

	convey.Convey("TestGetDatabaseSetDataBase", t, func() {
		SetDataBase(dbmock)
		db := GetDataBase()
		convey.So(db, convey.ShouldEqual, dbmock)
	})
}

func TestCheckDB(t *testing.T) {

	monkey.Patch(dbaccessor.CheckDataBase, func(Db dbaccessor.DbAccessor) error {
		return nil
	})
	defer monkey.UnpatchAll()
	monkey.Patch(GetClusterUUID, func() string {
		return "1111"
	})

	convey.Convey("TestCheckDB", t, func() {
		err := CheckDB()
		convey.So(err, convey.ShouldBeNil)
	})
}
