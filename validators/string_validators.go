package validators

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

// StringIsOneOf creates a validator for Terraform schema values that ensures the supplied value is one of the specified strings.
func StringIsOneOf(valueDescription string, permittedValues ...string) schema.SchemaValidateFunc {
	permittedValuesDescription := strings.Join(permittedValues, ", ")

	return func(value interface{}, key string) (warnings []string, errors []error) {
		stringValue := value.(string)

		for _, permittedValue := range permittedValues {
			if stringValue == permittedValue {
				return
			}
		}

		errors = append(errors, fmt.Errorf("invalid %s '%s' (valid values are [%s])",
			valueDescription,
			stringValue,
			permittedValuesDescription,
		))

		return
	}
}

// StringIsOneOfCaseInsensitive creates a validator for Terraform schema values that ensures the supplied value is one of the specified strings (case-insensitive).
func StringIsOneOfCaseInsensitive(valueDescription string, permittedValues ...string) schema.SchemaValidateFunc {
	upperPermittedValues := make([]string, len(permittedValues))
	for index, permittedValue := range permittedValues {
		upperPermittedValues[index] = strings.ToUpper(permittedValue)
	}

	permittedValuesDescription := strings.Join(permittedValues, ", ")

	return func(value interface{}, key string) (warnings []string, errors []error) {
		stringValue := value.(string)
		upperStringValue := strings.ToUpper(stringValue)

		for _, upperPermittedValue := range upperPermittedValues {
			if upperStringValue == upperPermittedValue {
				return
			}
		}

		errors = append(errors, fmt.Errorf("invalid %s '%s' (valid values are [%s])",
			valueDescription,
			stringValue,
			permittedValuesDescription,
		))

		return
	}
}
