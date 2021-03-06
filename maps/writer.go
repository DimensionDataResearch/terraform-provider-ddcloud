package maps

// Writer provides an interface for writing to a Terraform value map.
//
// ToSether, Reader and Writer provider a consistent interface.
type Writer interface {
	// SetString creates or updates a string in the underlying data.
	SetString(key string, value string)

	// SetStringPtr creates or updates a string pointer in the underlying data.
	SetStringPtr(key string, value *string)

	// SetStringSlice creates or updates a string slice in the underlying data.
	SetStringSlice(key string, value ...string)

	// SetInt creates or updates an integer in the underlying data.
	SetInt(key string, value int)

	// SetIntPtr creates or updates an integer pointer in the underlying data.
	SetIntPtr(key string, value *int)

	// SetMapSlice creates or updates a slice of maps in the underlying data.
	SetMapSlice(key string, values ...map[string]interface{})
}

// NewWriter creates a new Writer to read the values in the specified map.
func NewWriter(data map[string]interface{}) Writer {
	return &mapAdapter{
		data: data,
	}
}
