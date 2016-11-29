package maps

// Writer provides an interface for writing to a Terraform value map.
//
// ToSether, Reader and Writer provider a consistent interface.
type Writer interface {
	// SetString creates or updates a string in the underlying data.
	SetString(key string, value string)

	// SetStringPtr creates or updates a string pointer in the underlying data.
	SetStringPtr(key string, value *string)

	// SetInt creates or updates an integer in the underlying data.
	SetInt(key string, value int)

	// SetIntPtr creates or updates an integer pointer in the underlying data.
	SetIntPtr(key string, value *int)
}

// NewWriter creates a new Writer to read the values in the specified map.
func NewWriter(data map[string]interface{}) Writer {
	return &mapAdapter{
		data: data,
	}
}
