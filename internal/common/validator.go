package common

import (
	"regexp"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	FunctionNodeNameValidators []validator.String = []validator.String{
		stringvalidator.LengthBetween(3, 80),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[A-Za-z0-9\-\_ ]*$`),
			"value must contain only lowercase/uppercase alphanumeric characters, \"-\", or \"_\"",
		),
	}
	LogLevelValidator validator.String = stringvalidator.OneOf(
		string(api.LogLevelDebug),
		string(api.LogLevelError),
		string(api.LogLevelInfo),
		string(api.LogLevelWarning),
	)
	NameValidators []validator.String = []validator.String{
		stringvalidator.LengthBetween(3, 80),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[A-Za-z0-9\-\_\.\: ]*$`),
			"value must contain only lowercase/uppercase alphanumeric characters, \"-\", \"_\", \":\", or \".\"",
		),
	}
	PortValidator     validator.Int64  = int64validator.Between(1024, 65535)
	ProtocolValidator validator.String = stringvalidator.OneOf(
		string(api.ProtocolSctp),
		string(api.ProtocolTcp),
		string(api.ProtocolUdp),
	)
	RequirementsValidator validator.Set = setvalidator.ValueStringsAre(
		stringvalidator.LengthAtLeast(1),
	)
	SystemNameValidator validator.String = stringvalidator.RegexMatches(
		regexp.MustCompile(`^echo\..*$`),
		"value must begin with \"echo.\"",
	)
	NotSystemNameValidator validator.String = validators.Not(SystemNameValidator)
)
