package jsonify

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type Jsonifable interface {
	Jsonify() []byte
}

// JsonifySLn - makes json from val. Includes privat fields.
// use `jignore:""` tag for ignore field
// use `masker:""` tag for masking field with * (len will be same)
// use `masker:"fl"` tag for masking field exepting first and last value
// there is no cyclomatic protection
func JsonifySLn(val any) string {
	return string(Jsonify(val, " ", "\n"))
}

// JsonifySM - (single line) makes json from val. Includes privat fields.
// use `jignore:""` tag for ignore field
// use `masker:""` tag for masking field with * (len will be same)
// use `masker:"fl"` tag for masking field exepting first and last value
// there is no cyclomatic protection
func JsonifySM(val any) string {
	return string(Jsonify(val, "", ""))
}

// JsonifyS - makes json from val. Includes privat fields.
// use `jignore:""` tag for ignore field
// use `masker:""` tag for masking field with * (len will be same)
// use `masker:"fl"` tag for masking field exepting first and last value
// there is no cyclomatic protection
func JsonifyS(val any, indent string, nextRow string) string {
	return string(Jsonify(val, indent, nextRow))
}

// Jsonify - makes json from val. Includes privat fields.
// use `jignore:""` tag for ignore field
// use `masker:""` tag for masking field with * (len will be same)
// use `masker:"fl"` tag for masking field exepting first and last value
// there is no cyclomatic protection
func Jsonify(val any, indent string, nextRow string) []byte {
	if val == nil {
		return []byte("null")
	}

	return jsonify(reflect.ValueOf(val), "", "", jsonifyParams{
		indent:  indent,
		nextRow: nextRow,
	})

}

// JsonifyM - makes json from val. Includes privat fields.
// use `jignore:""` tag for ignore field
// use `masker:""` tag for masking field with * (len will be same)
// use `masker:"fl"` tag for masking field exepting first and last value
// there is no cyclomatic protection
func JsonifyM(val any) json.RawMessage {
	return Jsonify(val, "", "")
}

type jsonifyParams struct {
	indent  string
	nextRow string
}

func jsonify(r reflect.Value, indentNow string, typeNow string, params jsonifyParams) []byte {
	for k, v := range specMarchallerInterface {
		if r.Type().Implements(v) {
			mrs, ok := specMarchaller[k]
			if ok {
				res, ok := mrs(r, indentNow, typeNow, params.indent, params.nextRow)
				if ok {
					return res
				}
			}
		}
	}

	if SpecialMarchaller != nil {
		res, ok := SpecialMarchaller(r, indentNow, typeNow, params.indent, params.nextRow)
		if ok {
			return res
		}
	}

	mrs, ok := specMarchaller[r.Type().PkgPath()+"/"+r.Type().Name()]
	if ok {
		res, ok := mrs(r, indentNow, typeNow, params.indent, params.nextRow)
		if ok {
			return res
		}
	}

	mrs, ok = specMarchaller[r.Type().String()]
	if ok {
		res, ok := mrs(r, indentNow, typeNow, params.indent, params.nextRow)
		if ok {
			return res
		}
	}

	switch r.Kind() {
	case reflect.Bool:
		if r.Bool() {
			return []byte("true")
		}
		return []byte("false")
	case reflect.String:
		b, _ := json.Marshal(r.String())
		return b
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.Itoa(int(r.Int())))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.Itoa(int(r.Uint())))
	case reflect.Float32, reflect.Float64:
		return []byte(strconv.FormatFloat(r.Float(), 'f', -1, 64))
	case reflect.Complex64, reflect.Complex128:
		b, _ := json.Marshal(r.Complex())
		return b
	case reflect.Struct:
		return jsonifyStruct(r, indentNow, typeNow, params)
	case reflect.Slice:
		return jsonifySlice(r, indentNow, typeNow, params)
	case reflect.Map:
		if r.IsNil() {
			return []byte("null")
		}
		return jsonifyMap(r, indentNow, typeNow, params)
	case reflect.Array:
		return jsonifySlice(r, indentNow, typeNow, params)
	case reflect.Invalid:
		return []byte("\"reflect.Invalid\"")
	case reflect.Chan:
		return []byte("\"reflect.Chan\"")
	case reflect.Func:
		return []byte("\"reflect.Func\"")
	case reflect.Pointer, reflect.Interface:
		if r.IsNil() {
			return []byte("null")
		}
		i := r.Elem()
		return jsonify(i, indentNow, typeNow+"*", params)
	case reflect.UnsafePointer:
		return []byte("\"reflect.UnsafePointer\"")
	default:
		return []byte("\"" + fmt.Sprintf("unhandled kind %s", r.Kind()) + "\"")
	}
}

func jsonifyStruct(r reflect.Value, indentNow string, typeNow string, params jsonifyParams) []byte {
	res := []byte("{")

	skF := 0
	for i := 0; i < r.Type().NumField(); i++ {
		f := r.Type().Field(i)
		skip := false
		for tName, fnc := range ignoreTags {
			val, ok := f.Tag.Lookup(tName)
			if ok && fnc(val) {
				skF++
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if i > skF {
			res = append(res, []byte(",")...)
		}
		res = append(res, []byte(params.nextRow+indentNow+params.indent)...)
		res = append(res, jsonify(reflect.ValueOf(f.Name), indentNow+params.indent, "", params)...)
		res = append(res, []byte(": ")...)

		skip = false
		for tName, fnc := range maskTags {
			val, ok := f.Tag.Lookup(tName)
			if ok {
				bodyField, ok := fnc(val, r.FieldByName(f.Name), indentNow+params.indent, "", params.indent, params.nextRow)
				if ok {
					res = append(res, bodyField...)
					skip = true
					continue
				}
			}
		}
		if skip {
			continue
		}
		res = append(res, jsonify(r.FieldByName(f.Name), indentNow+params.indent, "", params)...)
	}

	res = append(res, []byte(params.nextRow+indentNow+"}")...)

	return res
}

func jsonifySlice(r reflect.Value, indentNow string, typeNow string, params jsonifyParams) []byte {
	res := []byte("[")

	for i := 0; i < r.Len(); i++ {
		f := r.Index(i)
		if i > 0 {
			res = append(res, []byte(",")...)
		}
		res = append(res, []byte(params.nextRow+indentNow+params.indent)...)
		res = append(res, jsonify(f, indentNow+params.indent, "", params)...)
	}

	res = append(res, []byte(params.nextRow+indentNow+"]")...)

	return res
}

func jsonifyMap(r reflect.Value, indentNow string, typeNow string, params jsonifyParams) []byte {
	res := []byte("{")

	iter := r.MapRange()
	i := 0
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		if i > 0 {
			res = append(res, []byte(",")...)
		}
		res = append(res, []byte(params.nextRow+indentNow+params.indent)...)
		res = append(res, jsonify(reflect.ValueOf(fmt.Sprint(k.String())), indentNow+params.indent, "", params)...)
		res = append(res, []byte(": ")...)
		res = append(res, jsonify(v, indentNow+params.indent, "", params)...)
		i++
	}

	res = append(res, []byte(params.nextRow+indentNow+"}")...)

	return res
}
