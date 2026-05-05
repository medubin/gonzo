package url

import (
	"context"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type ParamKey = struct{}

type fieldMeta struct {
	index int
	tag   string
}

type cacheKey struct {
	typ     reflect.Type
	tagName string
}

var fieldCache sync.Map // map[cacheKey][]fieldMeta

func getFieldMeta(t reflect.Type, tagName string) []fieldMeta {
	key := cacheKey{t, tagName}
	if v, ok := fieldCache.Load(key); ok {
		return v.([]fieldMeta)
	}
	var meta []fieldMeta
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get(tagName)
		if tag == "" {
			continue
		}
		if tagName == "json" {
			tag = strings.Split(tag, ",")[0]
		}
		if tag == "" || tag == "-" {
			continue
		}
		meta = append(meta, fieldMeta{index: i, tag: tag})
	}
	fieldCache.Store(key, meta)
	return meta
}

func ConvertPathToRegex(path string) *regexp.Regexp {
	path = strings.TrimRight(path, "/")
	// First convert all parameters that are not given an explicit match type. Give it \w+
	unspecifiedConversion := regexp.MustCompile(`{([a-zA-Z0-9]+)}`)
	path = unspecifiedConversion.ReplaceAllString(path, `(?P<$1>\w+)`)

	specifiedConversion := regexp.MustCompile(`{([a-zA-Z0-9]+):(.+)}`)
	path = specifiedConversion.ReplaceAllString(path, `(?P<$1>$2)`)

	// NOTE: ^ means start of string and $ means end. Without these,
	//   we'll still match if the path has content before or after
	//   the expression (/foo/bar/baz would match the "/bar" route).
	return regexp.MustCompile("^" + path + "/?$")
}

func GetTypedParamsFromContext[Params any](ctx context.Context) Params {
	var params Params

	paramMap, ok := ctx.Value(ParamKey{}).(map[string]string)
	if !ok {
		return params
	}

	setFieldsFromMap(&params, paramMap, "url")
	return params
}

func GetTypedParamsFromQuery[Params any](query url.Values) Params {
	var params Params
	setFieldsFromValues(&params, query, "json")
	return params
}

// setFieldsFromValues populates struct fields from url.Values. Slice fields
// (and *[]T pointer-to-slice fields) collect every value supplied for the tag,
// so `?tag=a&tag=b` produces []string{"a","b"} instead of dropping "b". Scalar
// fields take the first value, matching the previous behavior.
func setFieldsFromValues(structPtr interface{}, values url.Values, tagName string) {
	v := reflect.ValueOf(structPtr).Elem()
	for _, m := range getFieldMeta(v.Type(), tagName) {
		raw, exists := values[m.tag]
		if !exists || len(raw) == 0 {
			continue
		}
		field := v.Field(m.index)
		if !field.CanSet() {
			continue
		}
		if isSliceField(field.Type()) {
			setSliceField(field, raw)
		} else if !setFieldValue(field, raw[0]) {
			// Conversion failure used to silently leave the field at its
			// zero value, which made misformatted query/path params
			// indistinguishable from missing ones. Log a warning so the
			// failure is at least observable; the field still ends up zero
			// for backwards compatibility.
			log.Printf("gonzo: param %q: cannot convert %q to %s; field left at zero value", m.tag, raw[0], field.Type())
		}
	}
}

func isSliceField(t reflect.Type) bool {
	if t.Kind() == reflect.Slice {
		return true
	}
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Slice
}

// setSliceField fills a slice field (or *[]T) by converting each raw string
// with setScalarValue. Values that fail to convert are skipped, matching the
// scalar-field policy in setScalarValue.
func setSliceField(field reflect.Value, raw []string) {
	sliceType := field.Type()
	if sliceType.Kind() == reflect.Ptr {
		sliceType = sliceType.Elem()
	}
	elemType := sliceType.Elem()

	out := reflect.MakeSlice(sliceType, 0, len(raw))
	for _, s := range raw {
		elem := reflect.New(elemType).Elem()
		if setScalarValue(elem, s) {
			out = reflect.Append(out, elem)
		}
	}

	if field.Type().Kind() == reflect.Ptr {
		ptr := reflect.New(sliceType)
		ptr.Elem().Set(out)
		field.Set(ptr)
	} else {
		field.Set(out)
	}
}

// setFieldsFromMap uses reflection to set struct fields based on tag values.
// Field metadata is cached per (type, tagName) pair via sync.Map so the
// reflection loop over struct fields only runs once per type.
func setFieldsFromMap(structPtr interface{}, paramMap map[string]string, tagName string) {
	v := reflect.ValueOf(structPtr).Elem()
	for _, m := range getFieldMeta(v.Type(), tagName) {
		if value, exists := paramMap[m.tag]; exists {
			if !setFieldValue(v.Field(m.index), value) {
				log.Printf("gonzo: param %q: cannot convert %q to %s; field left at zero value", m.tag, value, v.Field(m.index).Type())
			}
		}
	}
}

// setFieldValue sets a struct field value from a string, handling type
// conversion. Returns true on success; false when the field can't be set or
// the value can't be parsed into the field's type. On failure the field is
// left at its zero value, matching pre-existing behavior.
func setFieldValue(field reflect.Value, value string) bool {
	if !field.CanSet() {
		return false
	}

	fieldType := field.Type()

	// Handle pointer types
	if fieldType.Kind() == reflect.Ptr {
		elemType := fieldType.Elem()
		newValue := reflect.New(elemType)
		if setScalarValue(newValue.Elem(), value) {
			field.Set(newValue)
			return true
		}
		return false
	}
	return setScalarValue(field, value)
}

// setScalarValue sets a non-pointer field value from a string
func setScalarValue(field reflect.Value, value string) bool {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		return true
	case reflect.Int32:
		if i, err := strconv.ParseInt(value, 10, 32); err == nil {
			field.SetInt(i)
			return true
		}
	case reflect.Int64:
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(i)
			return true
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(value); err == nil {
			field.SetBool(b)
			return true
		}
	case reflect.Float32:
		if f, err := strconv.ParseFloat(value, 32); err == nil {
			field.SetFloat(f)
			return true
		}
	case reflect.Float64:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(f)
			return true
		}
	}
	return false
}
