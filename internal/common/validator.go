package common

import (
	"regexp"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	LogLevelValidator tfsdk.AttributeValidator = stringvalidator.OneOf(
		string(api.LogLevelDebug),
		string(api.LogLevelError),
		string(api.LogLevelInfo),
		string(api.LogLevelWarning),
	)
	NameValidators []tfsdk.AttributeValidator = []tfsdk.AttributeValidator{
		stringvalidator.LengthBetween(3, 80),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[A-Za-z0-9\-\_\.\: ]*$`),
			"value must contain only lowercase/uppercase alphanumeric characters, \"-\", \"_\", \":\", or \".\"",
		),
	}
	RequirementsValidator tfsdk.AttributeValidator = setvalidator.ValuesAre(
		stringvalidator.LengthAtLeast(1),
	)
	SystemNameValidator tfsdk.AttributeValidator = stringvalidator.RegexMatches(
		regexp.MustCompile(`^echo\..*$`),
		"value must begin with \"echo.\"",
	)
	NotSystemNameValidator tfsdk.AttributeValidator = validators.Not(SystemNameValidator)
)
