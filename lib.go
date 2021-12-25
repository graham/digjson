package parsing

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func inner_dig(obj map[string]interface{}, rest string) (interface{}, bool) {
	chunks := strings.SplitN(rest, ".", 2)

	hit, found := obj[chunks[0]]

	if found == false {
		return nil, false
	}

	unboxed, ok := hit.(map[string]interface{})

	if !ok || len(chunks) == 1 {
		return hit, true
	} else {
		return inner_dig(unboxed, chunks[1])
	}

	return nil, false

}

func TransferMapToStruct(data map[string]interface{}, target interface{}) error {
	tType := reflect.TypeOf(target)
	tVal := reflect.ValueOf(target)

	for i := 0; i < tType.Elem().NumField(); i++ {
		field := tType.Elem().Field(i)
		key := field.Tag.Get("json")
		kind := field.Type.Kind()
		dValue, found := data[key]

		if !found {
			continue
		}

		//  Get reference of field value provided to input `d`
		result := tVal.Elem().Field(i)
		dType := reflect.TypeOf(dValue)

		if dType.Kind() == kind {
			result.Set(reflect.ValueOf(dValue))
		} else {
			convertedValue := reflect.ValueOf(dValue).Convert(field.Type)
			result.Set(convertedValue)
			return nil
		}
	}

	return nil
}

func TransferResultToSlice(data interface{}, target interface{}) error {
	if data == nil {
		return nil
	}

	unboxed, ok := data.(map[string]interface{})

	if ok {
		eleType := reflect.TypeOf(target).Elem().Elem()
		value := reflect.New(eleType)

		TransferMapToStruct(unboxed, value.Interface())

		array := reflect.ValueOf(target).Elem()
		result := reflect.Append(array, value.Elem())
		array.Set(result)
		return nil
	}

	unsliced, ok := data.([]interface{})

	if ok {
		for _, row := range unsliced {
			eleType := reflect.TypeOf(target).Elem().Elem()
			value := reflect.New(eleType)

			TransferMapToStruct(row.(map[string]interface{}), value.Interface())

			array := reflect.ValueOf(target).Elem()
			result := reflect.Append(array, value.Elem())
			array.Set(result)
		}
		return nil
	}

	return nil

}

func Dig(payload []byte, path string, target interface{}) (bool, error) {
	return dig(payload, path, target, false)
}

func DigStrict(payload []byte, path string, target interface{}) (bool, error) {
	return dig(payload, path, target, true)
}

func dig(payload []byte, path string, target interface{}, isStrict bool) (bool, error) {
	var t map[string]interface{}
	err := json.Unmarshal(payload, &t)

	newval, was_found := inner_dig(t, path)

	if was_found {
		targetType := reflect.Indirect(reflect.ValueOf(target))
		//valueType := reflect.TypeOf(newval)
		if targetType.Kind() == reflect.Struct {
			unboxed, ok := newval.(map[string]interface{})
			if !ok {
				return false, errors.New(fmt.Sprintf("Final obj wasn't map: %s", newval))
			}
			TransferMapToStruct(unboxed, target)
			return was_found, nil
		} else if targetType.Kind() == reflect.Slice {
			TransferResultToSlice(newval, target)
			return was_found, nil
		} else {
			indr := reflect.ValueOf(target)
			indr = reflect.Indirect(indr)

			fromType := reflect.TypeOf(newval)
			toType := reflect.TypeOf(target).Elem()

			if fromType.ConvertibleTo(toType) {
				newval := reflect.ValueOf(newval).Convert(toType)
				indr.Set(reflect.Indirect(newval))
				return was_found, nil
			}

			if isStrict {
				return was_found, errors.New(fmt.Sprintf("Types mismatch %s - %s (strict mode).", fromType, toType))
			}

			if fromType.Kind() == reflect.String && toType.Kind() == reflect.Int {
				intValue, err := strconv.ParseInt(newval.(string), 10, 64)
				if err != nil {
					return was_found, err
				}
				indr.SetInt(intValue)
				return was_found, nil
			}

			return was_found, errors.New(fmt.Sprintf("Can't convert %s to %s yet.", fromType, toType))

		}
	}

	return was_found, err
}
