package models

import (
	"encoding/json"
	"log"
	"reflect"
	"strings"
)

const NullString string = "__NULL__"

func (i IssueUpdateInput) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})
	val := reflect.ValueOf(&i).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Tag.Get("json")
		if commaIdx := len(fieldName); commaIdx > 0 {
			fieldName = fieldName[:commaIdx]
		}
		if fieldName == "-" {
			continue
		}

		fieldName = strings.Split(fieldName, ",")[0]

		if field.IsNil() {
			continue // omit nil fields unless they're explicitly set to the magic string
		}

		fieldValue := field.Interface()
		switch v := fieldValue.(type) {
		case *string:
			if v != nil && *v == NullString {
				result[fieldName] = nil
			} else {
				result[fieldName] = v
			}
		case *int64, *float64, *bool:
			result[fieldName] = v
		case []string:
			result[fieldName] = v
		default:
			result[fieldName] = v
		}
	}

	log.Printf("%+v", result)

	return json.Marshal(result)
}
