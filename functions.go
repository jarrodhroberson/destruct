package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"io"
	"reflect"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
)

func MapKeysAsSlice[K comparable, V any](m map[K]V) []K {
	ks := make([]K, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func includeFieldPredicate(f reflect.StructField) bool {
	if str := f.Tag.Get("identity"); str != "" {
		return str != "-"
	}
	return true
}

func primitiveStrategy(h io.Writer, rv reflect.Value) ([]byte, bool) {
	switch rv.Kind() {
	case reflect.String:
		return []byte(rv.String()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var b bytes.Buffer
		_ = binary.Write(&b, binary.BigEndian, rv.Int())
		return b.Bytes(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var b bytes.Buffer
		_ = binary.Write(&b, binary.BigEndian, rv.Uint())
		return b.Bytes(), true
	case reflect.Float32, reflect.Float64:
		var b bytes.Buffer
		_ = binary.Write(&b, binary.BigEndian, rv.Float())
		return b.Bytes(), true
	case reflect.Bool:
		var b bytes.Buffer
		_ = binary.Write(&b, binary.BigEndian, rv.Bool())
		return b.Bytes(), true
	default:
		return nil, false
	}
}

func pointerStrategy(h io.Writer, rv reflect.Value) ([]byte, bool) {
	if rv.Kind() == reflect.Ptr {
		if !rv.IsNil() || rv.Type().Elem().Kind() == reflect.Struct {
			return structStrategy(h, rv)
		} else {
			zero := reflect.Zero(rv.Type().Elem())
			if b, ok := primitiveStrategy(h, zero); ok {
				return b, true
			} else if b, ok := mapStrategy(h, zero); ok {
				return b, true
			} else if b, ok := pointerStrategy(h, zero); ok {
				return b, true
			}
			return nil, false
		}
	}
	return nil, false
}

func mapStrategy(h io.Writer, rv reflect.Value) ([]byte, bool) {
	if rv.Kind() == reflect.Map {
		mk := rv.MapKeys()
		kv := make(map[string]reflect.Value, len(mk))
		for _, k := range mk {
			kv[k.String()] = k
		}
		keys := MapKeysAsSlice[string, reflect.Value](kv)
		sort.Strings(keys)
		b := bytes.Buffer{}
		for idx := range keys {
			strategies{
				primitiveStrategy,
				pointerStrategy,
				mapStrategy,
				structStrategy,
				interfaceStrategy,
				arraySliceStrategy,
				defaultStrategy,
			}.apply(h, rv.MapIndex(kv[keys[idx]]))
		}
		return b.Bytes(), true
	}
	return nil, false
}

func timeStructStrategy(h io.Writer, v reflect.Value) ([]byte, bool) {
	if v.Kind() == reflect.Struct {
		log.Debug().Msgf("%v", v)
		if v.Type() == reflect.TypeOf(time.Time{}) {
			s := v.Interface().(time.Time).Format(time.RFC3339Nano)
			return []byte(s), true
		}
	}
	return nil, false
}

func structStrategy(h io.Writer, v reflect.Value) ([]byte, bool) {
	if v.Kind() == reflect.Struct {
		log.Debug().Msgf("%v", v)
		kv := make(map[string]reflect.Value, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			t1 := v.Field(i)
			if !includeFieldPredicate(v.Type().Field(i)) {
				continue
			}
			kv[t1.Type().Name()] = t1
		}

		keys := MapKeysAsSlice[string, reflect.Value](kv)
		sort.Strings(keys)
		b := bytes.Buffer{}
		s := strategies{
			primitiveStrategy,
			pointerStrategy,
			mapStrategy,
			timeStructStrategy,
			structStrategy,
			interfaceStrategy,
			arraySliceStrategy,
			defaultStrategy,
		}
		for _, key := range keys {
			value := kv[key]
			log.Debug().Msg(key)
			b.Write(s.apply(h, value))
		}
		return b.Bytes(), true
	}
	return nil, false
}

func interfaceStrategy(h io.Writer, rv reflect.Value) ([]byte, bool) {
	if rv.Kind() == reflect.Interface {
		if !rv.CanInterface() {
			return nil, false
		}
		strategies{
			primitiveStrategy,
			pointerStrategy,
			mapStrategy,
			timeStructStrategy,
			structStrategy,
			interfaceStrategy,
			arraySliceStrategy,
			defaultStrategy,
		}.apply(h, reflect.ValueOf(rv.Interface()))
	}
	return nil, false
}

func arraySliceStrategy(h io.Writer, rv reflect.Value) ([]byte, bool) {
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		var b bytes.Buffer
		for i := 0; i < rv.Len(); i++ {
			strategies{
				primitiveStrategy,
				timeStructStrategy,
				pointerStrategy,
				mapStrategy,
				structStrategy,
				interfaceStrategy,
				arraySliceStrategy,
				defaultStrategy,
			}.apply(h, rv)
		}
		return b.Bytes(), true
	default:
		return nil, false
	}
}

func defaultStrategy(h io.Writer, rv reflect.Value) ([]byte, bool) {
	return rv.Bytes(), true
}

type strategies []func(w io.Writer, rv reflect.Value) ([]byte, bool)

var identityStrategies = strategies{
	primitiveStrategy,
	pointerStrategy,
	mapStrategy,
	timeStructStrategy,
	structStrategy,
	interfaceStrategy,
	arraySliceStrategy,
	defaultStrategy,
}

func (is strategies) apply(h io.Writer, object any) []byte {
	for _, strategy := range is {
		if reflect.TypeOf(object) == reflect.TypeOf(reflect.Value{}) {
			if b, ok := strategy(h, object.(reflect.Value)); ok {
				return b
			}
		} else {
			if b, ok := strategy(h, reflect.ValueOf(object)); ok {
				return b
			}
		}
	}
	return []byte{}
}

func HashIdentity[T any](t T) string {
	h := sha512.New()
	i := identityStrategies.apply(h, t)
	s := hex.EncodeToString(i)
	return s
}
