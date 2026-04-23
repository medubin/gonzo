package form

import (
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/medubin/gonzo/runtime/gerrors"
)

const maxFormMemory = 32 << 20 // 32MB

var fileHeaderType = reflect.TypeOf((*multipart.FileHeader)(nil))

// Parse parses a multipart/form-data request into a struct of type T.
// Fields with `form:"name"` tags are populated from form values.
// Fields of type *multipart.FileHeader are populated from uploaded files.
func Parse[T any](r *http.Request) (*T, error) {
	if err := r.ParseMultipartForm(maxFormMemory); err != nil {
		return nil, gerrors.MalformedError("failed to parse multipart form: " + err.Error())
	}

	var result T
	v := reflect.ValueOf(&result).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		formTag := field.Tag.Get("form")
		if formTag == "" || formTag == "-" {
			continue
		}
		name, _, _ := strings.Cut(formTag, ",")

		if field.Type == fileHeaderType {
			_, fh, err := r.FormFile(name)
			if err == nil {
				fieldVal.Set(reflect.ValueOf(fh))
			} else if err != http.ErrMissingFile {
				return nil, gerrors.MalformedError("failed to read file '" + name + "': " + err.Error())
			}
			continue
		}

		if field.Type.Kind() == reflect.Ptr {
			val := r.FormValue(name)
			if val == "" {
				continue
			}
			ptr := reflect.New(field.Type.Elem())
			if err := setValue(ptr.Elem(), val); err != nil {
				return nil, err
			}
			fieldVal.Set(ptr)
		}
	}

	return &result, nil
}

func setValue(v reflect.Value, s string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return gerrors.InvalidArgumentError("invalid integer value: " + s)
		}
		v.SetInt(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return gerrors.InvalidArgumentError("invalid float value: " + s)
		}
		v.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return gerrors.InvalidArgumentError("invalid bool value: " + s)
		}
		v.SetBool(b)
	}
	return nil
}
