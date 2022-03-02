package utils

import (
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"
)

func IsEmptyInsert(rst *mongo.SingleResult) bool {
	v := reflect.ValueOf(rst).FieldByName("cur")
	return !v.IsNil()
}
