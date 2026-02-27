package filter

import "reflect"

func (f Filter) Struct(obj any) (Fields, error) {
	if obj == nil {
		return nil, ErrNilData
	}

	// using reflect, we'll go through a struct and parse all its fields
	t := reflect.TypeOf(obj)
	if typeIsPointer(t) {
		t = t.Elem()
	}

	// if it's still a pointer, we're not playing that game
	if typeIsPointer(t) {
		return nil, ErrMultipleIndirects
	}

	// make sure it's a struct
	if !typeIsStruct(t) {
		return nil, ErrInvalidDataType
	}

	fields := make(Fields)

	next := []string{}
	nextMap := map[string]reflect.StructField{}

	for field := range t.Fields() {
		if !field.IsExported() && f.config.SkipHiddenFields {
			continue
		}

		// get the field type, and if it's a pointer, we'll get the element type
		// TODO: this doesn't really support mutliple levels of inderection yet,
		// which I think we want to let people do. Could I have just written it
		// in the time it took to write this comment? We'll never know.
		fieldType := field.Type
		if fieldIsPointer(field) {
			fieldType = fieldType.Elem()
		}

		// if the field is a struct, we'll add it to the queue to be processed later
		if typeIsStruct(fieldType) {
			next = append(next, field.Name)
			nextMap[field.Name] = field
			continue
		}

		fields[field.Name] = mapFieldType(fieldType) // we'll see in testing if this is a bad idea
	}

	// now we'll process the queue of struct fields in a cool psycho-loop style
	// no recursion for reasons of trauma
	for i := 0; i < len(next); i++ {
		field := nextMap[next[i]]

		// we're just doing the same logic as above - it's kind of like cheating!
		// not DRY, you say? Read closer, I say!
		fieldType := field.Type
		if fieldIsPointer(field) {
			fieldType = fieldType.Elem()
		}

		// TODO: support multiple indirects

		for subField := range fieldType.Fields() {
			if !subField.IsExported() && f.config.SkipHiddenFields {
				continue
			}

			subFieldType := subField.Type
			if fieldIsPointer(subField) {
				subFieldType = subFieldType.Elem()
			}

			if typeIsStruct(subFieldType) {
				next = append(next, field.Name+"."+subField.Name)
				nextMap[field.Name+"."+subField.Name] = subField
				continue
			}

			fields[field.Name+"."+subField.Name] = mapFieldType(subFieldType)
		}
	}

	return fields, nil
}

func typeIsPointer(t reflect.Type) bool {
	return t.Kind() == reflect.Pointer || t.Kind() == reflect.Ptr
}

func typeIsStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func fieldIsPointer(field reflect.StructField) bool {
	return field.Type.Kind() == reflect.Pointer || field.Type.Kind() == reflect.Ptr
}

func mapFieldType(t reflect.Type) fieldType {
	switch t.Kind() {
	case reflect.Bool:
		return TypeBool
	case reflect.String:
		return TypeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return TypeNumber
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return TypeBytes
		}
	}

	panic("unsupported field type: " + t.String())
}
