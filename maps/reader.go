package maps

// Reader provides an interface for reading from a Terraform value map.
//
// Together, Reader and Writer provider a consistent interface.
type Reader interface {
	// GetString retrieves a string from the underlying data.
	//
	// If the value does not exist, or is not a string, returns an empty string.
	GetString(key string) string

	// GetStringPtr retrieves a string pointer from the underlying data.
	//
	// If the value does not exist, or is not a string, returns nil.
	GetStringPtr(key string) *string

	// GetStringSlice retrieves a slice of strings from the underlying data, or an empty slice if not present.
	GetStringSlice(key string) []string

	// GetInt retrieves an integer from the underlying data.
	//
	// If the value does not exist, or is not an int, returns 0.
	GetInt(key string) int

	// GetInt retrieves an integer from the underlying data, or a default value if not present.
	//
	// If the value does not exist, or is not an int, returns defaultValue.
	GetIntOr(key string, defaultValue int) int

	// GetIntPtr retrieves an integer pointer from the underlying data.
	//
	// If the value does not exist, or is not an int, returns nil.
	GetIntPtr(key string) *int

	// GetMapSlice retrieves a slice of sub-maps from the underlying data.
	//
	// If the value does not exist, or is not a slice of maps, returns an empty slice.
	GetMapSlice(key string) []map[string]interface{}

	// GetMapSliceElement retrieves the specified element from a slice of sub-maps.
	//
	// If the value does not exist, or is not a slice of maps, returns nil.
	GetMapSliceElement(key string, index int) map[string]interface{}
}

// NewReader creates a new Reader to read the values in the specified map.
func NewReader(data map[string]interface{}) Reader {
	return &mapAdapter{
		data: data,
	}
}
