package maps

// mapAdapter is a simple implementation of an
type mapAdapter struct {
	data map[string]interface{}
}

// GetString retrieves a string from the underlying data.
//
// If the value does not exist, or is not a string, returns an empty string.
func (readerWriter *mapAdapter) GetString(key string) string {
	value, ok := readerWriter.data[key]
	if !ok {
		return ""
	}

	stringValue, ok := value.(string)
	if !ok {
		return ""
	}

	return stringValue
}

// GetStringPtr retrieves a string pointer from the underlying data.
//
// If the value does not exist, or is not a string, returns nil.
func (readerWriter *mapAdapter) GetStringPtr(key string) *string {
	value, ok := readerWriter.data[key]
	if !ok {
		return nil
	}

	stringPtrValue, ok := value.(*string)
	if ok {
		return stringPtrValue
	}

	stringValue, ok := value.(string)
	if ok {
		return &stringValue
	}

	return nil
}

// GetStringSlice retrieves a slice of strings from the underlying data, or an empty slice if the value does not exist.
func (readerWriter *mapAdapter) GetStringSlice(key string) []string {
	strings := make([]string, 0)

	value, ok := readerWriter.data[key]
	if !ok {
		return strings
	}

	rawSlice, ok := value.([]interface{})
	if !ok {
		return strings
	}

	for index := range rawSlice {
		stringValue, ok := rawSlice[index].(string)
		if ok {
			strings = append(strings, stringValue)
		}
	}

	return strings
}

// GetInt retrieves an integer from the underlying data.
//
// If the value does not exist, or is not an int, returns 0.
func (readerWriter *mapAdapter) GetInt(key string) int {
	return readerWriter.GetIntOr(key, 0)
}

// GetInt retrieves an integer from the underlying data, or a default value if not present.
//
// If the value does not exist, or is not an int, returns defaultValue.
func (readerWriter *mapAdapter) GetIntOr(key string, defaultValue int) int {
	value, ok := readerWriter.data[key]
	if !ok {
		return defaultValue
	}

	intValue, ok := value.(int)
	if !ok {
		return defaultValue
	}

	return intValue
}

// GetIntPtr retrieves an integer pointer from the underlying data.
//
// If the value does not exist, or is not an int, returns nil.
func (readerWriter *mapAdapter) GetIntPtr(key string) *int {
	value, ok := readerWriter.data[key]
	if !ok {
		return nil
	}

	intPtrValue, ok := value.(*int)
	if ok {
		return intPtrValue
	}

	intValue, ok := value.(int)
	if ok {
		return &intValue
	}

	return nil
}

// GetMapSlice retrieves a slice of sub-maps from the underlying data.
//
// If the value does not exist, or is not a slice of maps, returns an empty slice.
func (readerWriter *mapAdapter) GetMapSlice(key string) []map[string]interface{} {
	value, ok := readerWriter.data[key]
	if !ok {
		return make([]map[string]interface{}, 0)
	}

	slice, ok := value.([]interface{})
	if !ok {
		return make([]map[string]interface{}, 0)
	}

	maps := make([]map[string]interface{}, len(slice))
	for index := range slice {
		maps[index], ok = slice[index].(map[string]interface{})
		if !ok {
			maps[index] = make(map[string]interface{}) // More forgiving behaviour: "not a map" becomes "empty map".
		}
	}

	return maps
}

// GetMapSliceElement retrieves the specified element from a slice of sub-maps in the underlying data.
//
// If the value does not exist, or is not a slice of maps, returns nil.
func (readerWriter *mapAdapter) GetMapSliceElement(key string, index int) map[string]interface{} {
	value, ok := readerWriter.data[key]
	if !ok {
		return nil
	}

	slice, ok := value.([]interface{})
	if !ok {
		return nil
	}
	if index >= len(slice) {
		return nil
	}

	element, ok := slice[index].(map[string]interface{})
	if !ok {
		return nil
	}

	return element
}

// SetString creates or updates a string in the underlying data.
func (readerWriter *mapAdapter) SetString(key string, value string) {
	readerWriter.data[key] = value
}

// SetStringPtr creates or updates a string pointer in the underlying data.
func (readerWriter *mapAdapter) SetStringPtr(key string, value *string) {
	readerWriter.data[key] = value
}

// SetStringSlice creates or updates a string in the underlying data.
func (readerWriter *mapAdapter) SetStringSlice(key string, values ...string) {
	strings := make([]interface{}, len(values))
	for index := range values {
		strings[index] = values[index]
	}

	readerWriter.data[key] = strings
}

// SetInt creates or updates an integer in the underlying data.
func (readerWriter *mapAdapter) SetInt(key string, value int) {
	readerWriter.data[key] = value
}

// SetIntPtr creates or updates an integer pointer in the underlying data.
func (readerWriter *mapAdapter) SetIntPtr(key string, value *int) {
	readerWriter.data[key] = value
}

// SetMapSlice creates or updates a slice of maps in the underlying data.
func (readerWriter *mapAdapter) SetMapSlice(key string, values ...map[string]interface{}) {
	slice := make([]interface{}, len(values))
	for index := range values {
		slice[index] = values[index]
	}

	readerWriter.data[key] = slice
}
