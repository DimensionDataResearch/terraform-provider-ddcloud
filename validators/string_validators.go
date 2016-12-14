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

		errors = append(errors, fmt.Errorf("Invalid %s '%s' (valid values are [%s])",
			valueDescription,
			value,
			permittedValuesDescription,
		))

		return
	}
}
