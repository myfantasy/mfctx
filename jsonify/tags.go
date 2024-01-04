package jsonify

import (
	"fmt"
	"reflect"
)

const IgnoreTag string = "jignore"
const MaskTag string = "masker"
const MaskTagFirstLast string = "fl"

type CheckIgnoreTagFunc func(tag string) bool
type MaskTagFunc func(tag string, r reflect.Value, indentNow string, typeNow string, indent string, nextRow string) (res []byte, ok bool)

var ignoreTags map[string]CheckIgnoreTagFunc = make(map[string]CheckIgnoreTagFunc, 0)
var maskTags map[string]MaskTagFunc = make(map[string]MaskTagFunc, 0)

func AddIgnoreTag(tagName string, checkFunc CheckIgnoreTagFunc) {
	ignoreTags[tagName] = checkFunc
}

func AddMaskTag(tagName string, maskFunc MaskTagFunc) {
	maskTags[tagName] = maskFunc
}

func init() {
	AddIgnoreTag(IgnoreTag, func(tag string) bool {
		return true
	})
	AddMaskTag(MaskTag, func(tag string, r reflect.Value, indentNow, typeNow, indent, nextRow string) (res []byte, ok bool) {
		s := fmt.Sprint(r)
		res = []byte("\"")
		for i, p := range s {
			if i == 0 {
				if tag == MaskTagFirstLast {
					res = append(res, []byte(string(p))...)
				} else {
					res = append(res, byte('*'))
				}
			} else if i == len(s)-1 {
				if tag == MaskTagFirstLast {
					res = append(res, []byte(string(p))...)
				} else {
					res = append(res, byte('*'))
				}
			} else {
				res = append(res, byte('*'))
			}
		}

		res = append(res, []byte("\"")...)
		return res, true
	})
}
