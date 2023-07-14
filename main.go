package main

import (
	"reflect"
	"time"
)

type TimeStamp struct {
	T time.Time
}

type structTagType struct {
	Type  reflect.Type
	Tag   reflect.StructTag
	Value interface{}
}

func attributes(m interface{}) map[string]structTagType {

	attrs := make(map[string]structTagType)

	v := reflect.ValueOf(m)
	for i := 0; i < v.NumField(); i++ {
		t1 := v.Field(i)
		var val interface{}

		switch t1.Kind() {
		case reflect.Struct:
			if t1.Type() == reflect.TypeOf(time.Time{}) {
				val = t1.Interface().(time.Time).Format(time.RFC3339Nano) // Error in .Interface()
			}
		}

		aa := structTagType{}
		aa.Type = v.Type().Field(i).Type
		aa.Tag = v.Type().Field(i).Tag
		aa.Value = val
		attrs[v.Type().Field(i).Name] = aa
	}
	return attrs
}

func main() {
	model := TimeStamp{
		T: time.Now(),
	}
	//for name, mtype := range attributes(model) {
	//	Property := &mtype
	//	fmt.Printf("%6s, value: %10s\n", name, Property.Value)
	//}
	HashIdentity(model)
}

//panic: reflect.Value.Interface: cannot return value obtained from unexported field or method
