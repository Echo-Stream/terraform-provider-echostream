package validators

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = ipaddrValidator{}

// notValidator validates that value does not validate against the value validator.
type ipaddrValidator struct {
}

// Description describes the validation in plain text formatting.
func (v ipaddrValidator) Description(ctx context.Context) string {
	return "Value must be a valid IPv4 or IPv6 address."
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v ipaddrValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
// The validator will pass if it encounters a value validator that returns no errors and will then return any warnings
// from the passing validator. Using All validator as value validators will pass if all the validators supplied in an
// All validator pass.
func (v ipaddrValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	s := req.ConfigValue.ValueString()

	if ip := net.ParseIP(s); ip == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Expected valid IPv4/IPv6 address",
			s,
		)
	}
}

// Any returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Validates against at least one of the value validators.
//
// To prevent practitioner confusion should non-passing validators have
// conflicting logic, only warnings from the passing validator are returned.
// Use AnyWithAllWarnings() to return warnings from non-passing validators
// as well.
func Ipaddr() validator.String {
	return ipaddrValidator{}
}
