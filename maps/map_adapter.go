package maps

// mapAdapter is a simple implementation of an
type mapAdapter struct {
	data map[string]interface{}
}

// GetString retrieves a string from the underlying data.
//
// If the value does not exist, or is not a string, returns an empty string.
func (reader *mapAdapter) GetString(key string) string {
	value, ok := reader.data[key]
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
func (reader *mapAdapter) GetStringPtr(key string) *string {
	value, ok := reader.data[key]
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

// GetInt retrieves an integer from the underlying data.
//
// If the value does not exist, or is not an int, returns 0.
func (reader *mapAdapter) GetInt(key string) int {
	return reader.GetIntOr(key, 0)
}

// GetInt retrieves an integer from the underlying data, or a default value if not present.
//
// If the value does not exist, or is not an int, returns defaultValue.
func (reader *mapAdapter) GetIntOr(key string, defaultValue int) int {
	value, ok := reader.data[key]
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
func (reader *mapAdapter) GetIntPtr(key string) *int {
	value, ok := reader.data[key]
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

// SetString creates or updates a string in the underlying data.
func (reader *mapAdapter) SetString(key string, value string) {
	reader.data[key] = value
}

// SetStringPtr creates or updates a string pointer in the underlying data.
func (reader *mapAdapter) SetStringPtr(key string, value *string) {
	reader.data[key] = value
}

// SetInt creates or updates an integer in the underlying data.
func (reader *mapAdapter) SetInt(key string, value int) {
	reader.data[key] = value
}

// SetIntPtr creates or updates an integer pointer in the underlying data.
func (reader *mapAdapter) SetIntPtr(key string, value *int) {
	reader.data[key] = value
}
