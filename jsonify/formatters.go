package jsonify

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/myfantasy/ints"
)

type JsonifyMarchaller func(r reflect.Value, indentNow string, typeNow string, indent string, nextRow string) (res []byte, ok bool)

var SpecialMarchaller JsonifyMarchaller

var specMarchaller map[string]JsonifyMarchaller = make(map[string]JsonifyMarchaller)
var specMarchallerInterface map[string]reflect.Type = make(map[string]reflect.Type)

func AddMarchallerByName(typeName string, marchaller JsonifyMarchaller) {
	specMarchaller[typeName] = marchaller
}

func AddMarchallerByExample(val any, marchaller JsonifyMarchaller) {
	r := reflect.ValueOf(val)
	AddMarchallerByName(r.Type().PkgPath()+"/"+r.Type().Name(), marchaller)
}

func AddMarchallerByExamplePartType(val any, marchaller JsonifyMarchaller) {
	r := reflect.ValueOf(val)
	AddMarchallerByName(r.Type().String(), marchaller)
}

func AddMarchallerByInterface(t reflect.Type, marchaller JsonifyMarchaller) {
	specMarchallerInterface[t.String()+"_int_"+t.PkgPath()+"/"+t.Name()] = t
	AddMarchallerByName(t.String()+"_int_"+t.PkgPath()+"/"+t.Name(), marchaller)
}

func GetSpecMarchaller() map[string]JsonifyMarchaller {
	return specMarchaller
}

func init() {
	AddMarchallerByExample(time.Time{}, func(r reflect.Value, indentNow, typeNow, indent, nextRow string) (res []byte, ok bool) {
		t := r.Interface().(time.Time)
		return []byte("\"" + t.Format(time.RFC3339Nano) + "\""), true
	})

	AddMarchallerByExample(ints.Uuid{}, func(r reflect.Value, indentNow, typeNow, indent, nextRow string) (res []byte, ok bool) {
		t := r.Interface().(ints.Uuid)
		return []byte("\"" + t.String() + "\""), true
	})

	AddMarchallerByExamplePartType(errors.New("abc"), func(r reflect.Value, indentNow, typeNow, indent, nextRow string) (res []byte, ok bool) {
		t := r.Interface().(error)
		json, _ := json.Marshal(t.Error())
		return json, true
	})

	AddMarchallerByExamplePartType(fmt.Errorf("abc %w", errors.New("abc")), func(r reflect.Value, indentNow, typeNow, indent, nextRow string) (res []byte, ok bool) {
		t := r.Interface().(error)
		json, _ := json.Marshal(t.Error())
		return json, true
	})

	AddMarchallerByInterface(reflect.TypeOf((*Jsonifable)(nil)).Elem(), func(r reflect.Value, indentNow, typeNow, indent, nextRow string) (res []byte, ok bool) {
		return r.Interface().(Jsonifable).Jsonify(), true
	})

	AddMarchallerByExamplePartType(json.RawMessage("rmsg"), func(r reflect.Value, indentNow, typeNow, indent, nextRow string) (res []byte, ok bool) {
		if r.IsNil() {
			t := json.RawMessage("null")
			return t, true
		}
		if !r.CanInterface() {
			t := r.Bytes()
			return t, true
		}
		t := r.Interface().(json.RawMessage)
		if t == nil {
			t = json.RawMessage("null")
		}
		return t, true
	})
}
