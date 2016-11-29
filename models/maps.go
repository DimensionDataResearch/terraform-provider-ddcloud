package models

type mapReader struct {
	data map[string]interface{}
}

func reader(data map[string]interface{}) *mapReader {
	return &mapReader{
		data: data,
	}
}

func (reader *mapReader) String(key string) string {
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
func (reader *mapReader) StringPtr(key string) *string {
	value, ok := reader.data[key]
	if !ok {
		return nil
	}

	stringValue, ok := value.(string)
	if !ok {
		return nil
	}

	return &stringValue
}

func (reader *mapReader) Int(key string) int {
	value, ok := reader.data[key]
	if !ok {
		return 0
	}

	intValue, ok := value.(int)
	if !ok {
		return 0
	}

	return intValue
}
func (reader *mapReader) IntPtr(key string) *int {
	value, ok := reader.data[key]
	if !ok {
		return nil
	}

	intValue, ok := value.(int)
	if !ok {
		return nil
	}

	return &intValue
}
