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

func primitiveStrategy(h io.Writer, rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.String:
		h.Write([]byte(rv.String()))
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var b bytes.Buffer
		_ = binary.Write(&b, binary.BigEndian, rv.Int())
		h.Write(b.Bytes())
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var b bytes.Buffer
		_ = binary.Write(&b, binary.BigEndian, rv.Uint())
		h.Write(b.Bytes())
		return true
	case reflect.Float32, reflect.Float64:
		var b bytes.Buffer
		_ = binary.Write(&b, binary.BigEndian, rv.Float())
		h.Write(b.Bytes())
		return true
	case reflect.Bool:
		var b bytes.Buffer
		_ = binary.Write(&b, binary.BigEndian, rv.Bool())
		h.Write(b.Bytes())
		return true
	default:
		return false
	}
}

func pointerStrategy(h io.Writer, rv reflect.Value) bool {
	if rv.Kind() == reflect.Ptr {
		if !rv.IsNil() || rv.Type().Elem().Kind() == reflect.Struct {
			return structStrategy(h, rv)
		} else {
			zero := reflect.Zero(rv.Type().Elem())
			if primitiveStrategy(h, zero) {
				return true
			} else if mapStrategy(h, zero) {
				return true
			} else if pointerStrategy(h, zero) {
				return true
			}
			return false
		}
	}
	return false
}

func mapStrategy(h io.Writer, rv reflect.Value) bool {
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
		h.Write(b.Bytes())
		return true
	}
	return false
}

func timeStructStrategy(h io.Writer, v reflect.Value) bool {
	if v.Kind() == reflect.Struct {
		log.Debug().Msgf("%v", v)
		if v.Type() == reflect.TypeOf(time.Time{}) {
			s := v.Interface().(time.Time).Format(time.RFC3339Nano)
			h.Write([]byte(s))
			return true
		}
	}
	return false
}

func structStrategy(h io.Writer, v reflect.Value) bool {
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
			s.apply(h, value)
		}
		return true
	}
	return false
}

func interfaceStrategy(h io.Writer, rv reflect.Value) bool {
	if rv.Kind() == reflect.Interface {
		if !rv.CanInterface() {
			return false
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
	return false
}

func arraySliceStrategy(h io.Writer, rv reflect.Value) bool {
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
		h.Write(b.Bytes())
		return true
	default:
		return false
	}
}

func defaultStrategy(h io.Writer, rv reflect.Value) bool {
	h.Write(rv.Bytes())
	return true
}

type strategies []func(w io.Writer, rv reflect.Value) bool

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

func (is strategies) apply(h io.Writer, object any) bool {
	for _, strategy := range is {
		if reflect.TypeOf(object) == reflect.TypeOf(reflect.Value{}) {
			if strategy(h, object.(reflect.Value)) {
				return true
			}
		} else {
			if strategy(h, reflect.ValueOf(object)) {
				return true
			}
		}
	}
	return false
}

func HashIdentity[T any](t T) string {
	h := sha512.New()
	identityStrategies.apply(h, t)
	s := hex.EncodeToString(h.Sum(nil))
	return s
}
