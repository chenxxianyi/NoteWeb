package models

import (
	"reflect"
	"strings"
	"testing"
)

func TestAnnotationPositionDataHasNoDefaultTag(t *testing.T) {
	field, ok := reflect.TypeOf(Annotation{}).FieldByName("PositionData")
	if !ok {
		t.Fatal("PositionData field not found")
	}

	tag := string(field.Tag.Get("gorm"))
	if strings.Contains(tag, "default") {
		t.Fatalf("PositionData uses a MySQL-incompatible default tag: %q", tag)
	}
}
