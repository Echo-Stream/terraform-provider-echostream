package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.AttributeValidator = notValidator{}

// notValidator validates that value does not validate against the value validator.
type notValidator struct {
	valueValidator tfsdk.AttributeValidator
}

// Description describes the validation in plain text formatting.
func (v notValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Value must not satisfy the validation: %s.", v.valueValidator.Description(ctx))
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v notValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
// The validator will pass if it encounters a value validator that returns no errors and will then return any warnings
// from the passing validator. Using All validator as value validators will pass if all the validators supplied in an
// All validator pass.
func (v notValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	validatorResp := &tfsdk.ValidateAttributeResponse{
		Diagnostics: diag.Diagnostics{},
	}

	v.valueValidator.Validate(ctx, req, validatorResp)

	// If there was an error then the not condition is true, simply return
	if validatorResp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(
		diag.NewAttributeErrorDiagnostic(
			req.AttributePath,
			"Invalid not condition",
			fmt.Sprintf("NOT %s", v.valueValidator.Description(ctx)),
		),
	)
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
func Not(valueValidator tfsdk.AttributeValidator) tfsdk.AttributeValidator {
	return notValidator{valueValidator: valueValidator}
}
