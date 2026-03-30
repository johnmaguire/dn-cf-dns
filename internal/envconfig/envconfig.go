// Package envconfig overlays environment variables onto a struct using
// `envconfig` struct tags. It supports string, bool, and []string (comma-separated) fields,
// and recursively processes embedded structs.
package envconfig

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Process reads environment variables into the fields of the struct pointed to
// by v, using the `envconfig` struct tag to determine the variable name. Fields
// without the tag are skipped. Nested structs are processed recursively.
func Process(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("envconfig: expected pointer to struct, got %T", v)
	}
	return processStruct(rv.Elem())
}

func processStruct(rv reflect.Value) error {
	rt := rv.Type()
	for i := range rt.NumField() {
		field := rt.Field(i)
		fv := rv.Field(i)

		// Recurse into nested structs
		if field.Type.Kind() == reflect.Struct {
			if err := processStruct(fv); err != nil {
				return err
			}
			continue
		}

		envVar := field.Tag.Get("envconfig")
		if envVar == "" {
			continue
		}

		val, ok := os.LookupEnv(envVar)
		if !ok {
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
			fv.SetString(val)
		case reflect.Bool:
			fv.SetBool(val == "1" || strings.EqualFold(val, "true"))
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				parts := strings.Split(val, ",")
				for i := range parts {
					parts[i] = strings.TrimSpace(parts[i])
				}
				fv.Set(reflect.ValueOf(parts))
			}
		}
	}
	return nil
}
