package digjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Used to dig into the json document and get to the value you're looking for.
// The assumption is that inner_dig is getting the standard map[string]interface{}
// that json.Unmarshal will return if asked (likely works with other json decoders)
// but I haven't tried it.
func inner_dig(obj map[string]interface{}, rest string) (interface{}, bool) {
	chunks := strings.SplitN(rest, ".", 2)

	hit, found := obj[chunks[0]]

	if found == false {
		return nil, false
	}

	// Is the value we found a {} or some other type.
	unboxed, ok := hit.(map[string]interface{})

	if !ok || len(chunks) == 1 {
		// If we didn't find a {} or we are out of path chunks
		// return what we have.
		return hit, true
	} else {
		// Otherwise keep digging!
		return inner_dig(unboxed, chunks[1])
	}

	// Tthe field we requested doesnt exist
	// stop digging and return nil
	return nil, false

}

// Since we've already decoded the json, we're long past the json module assigning
// fields with the `json` tags. This method transfers those directives to a struct.
// For now it doesn't support some of the more complicated directives, just the json name.
func TransferMapToStruct(data map[string]interface{}, target interface{}) error {
	tType := reflect.TypeOf(target)
	tVal := reflect.ValueOf(target)

	for i := 0; i < tType.Elem().NumField(); i++ {
		field := tType.Elem().Field(i)
		key := field.Tag.Get("json")
		kind := field.Type.Kind()
		dValue, found := data[key]

		if !found {
			// Ignore struct fields if the json
			// object we are looking at doesn't have
			// that key.
			continue
		}

		//  Get reference of field value provided to input `d`
		result := tVal.Elem().Field(i)
		dType := reflect.TypeOf(dValue)

		// If the kinds are the same, no conversion is needed
		// This is likely a direct memory copy.
		if dType.Kind() == kind {
			result.Set(reflect.ValueOf(dValue))
			continue
		}

		// If the kinds differ and Reflect cannot convert them for us,
		// bail out.
		if reflect.TypeOf(dValue).ConvertibleTo(field.Type) == false {
			return errors.New(fmt.Sprintf("cant convert %s to %s", reflect.ValueOf(dValue), field.Type))
		}

		// This should use transmute_value_to_target todo(graham)
		convertedValue := reflect.ValueOf(dValue).Convert(field.Type)
		result.Set(convertedValue)
	}

	return nil
}

// Similar to the above, if the user provided "target" for the data is a slice
// we have to generate the correct data and write it into the slice.
func TransferResultToSlice(data interface{}, target interface{}) error {
	// It is possible that the data path we used to dig into the
	// json document lead us to a null value, in that case.
	// In this case, don't touch the variable and return no error.
	// It might make more sense to clear the array in this case,
	// as someone could be re-using the slice, but for now we'll
	// just leave things alone.
	if data == nil {
		return nil
	}

	// Our dig path may have lead us to a {} (object), in that case,
	// We still want to write into the slice, but we want to write
	// one value into the list.
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

	// This is what nice APIs do, they will return a list no matter
	// what. In this case, we simply want to make sure we are creating
	// the right struct and copying the information into the correct
	// fields.
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

// Allow implicit conversions if the target struct has a field of a type that
// can be easily converted, for example, string to int is implemented,
// it would be trivial to implement more conversions.
func Dig(payload []byte, path string, target interface{}) (bool, error) {
	return dig(payload, path, target, false)
}

// Don't allow any conversions, if the source (json) type doesn't match the
// target (struct) type, bail out.
func DigStrict(payload []byte, path string, target interface{}) (bool, error) {
	return dig(payload, path, target, true)
}

// The most important part of the implementation, determine the source and target types
// determine if any conversion should happen, and make sure that it happens without error.
// Because in this case, your return value could be null as the "correct" value, it's
// important to know if the dig path actually resolved in the json document, or if some
// other issue occured. This is the bool returned from this method.
// (was_the_path_found, and error).
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
				// Easy conversion by the reflect module, likely more reliable as well.
				// This covers most of the int32 -> int64 or other safe operations.
				// I'm not sure how this handles overflow situations.
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

			err := transmute_value_to_target(fromType.Kind(), toType.Kind(), newval, &indr)

			if err != nil {
				return was_found, err
			}

			return was_found, nil

		}
	}

	return was_found, err
}

// I understand why this function will ruffle some feathers, changing values without the
// engineer being involved seems weird. That said, the source data is of one type,
// and the target struct likely is the datatype the engineer will want, we might as well
// make their life easier. It's a cost, but one I don't think is unreasonable.
func transmute_value_to_target(fromKind reflect.Kind, toKind reflect.Kind, source interface{}, target *reflect.Value) error {
	if fromKind == reflect.String && toKind == reflect.Int {
		intValue, err := strconv.ParseInt(source.(string), 10, 64)
		if err != nil {
			return err
		}
		target.SetInt(intValue)
		return nil
	} else if fromKind == reflect.Int && toKind == reflect.String {
		target.SetString(fmt.Sprintf("%d", source.(int64)))
	}

	return errors.New(fmt.Sprintf("Transmute from %s to %s not supported yet.", fromKind, toKind))
}
