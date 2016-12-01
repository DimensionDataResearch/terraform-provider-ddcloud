package models

func stringToPtr(value string) *string {
	if value != "" {
		return &value
	}

	return nil
}

func ptrToString(value *string) string {
	if value != nil {
		return *value
	}

	return ""
}
