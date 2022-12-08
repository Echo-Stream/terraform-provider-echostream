package common

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ types.StringTypable    = &ConfigType{}
	_ xattr.TypeWithValidate = &ConfigType{}
	_ types.StringValuable   = &Config{}
)

type ConfigType struct{}

func (ct ConfigType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, ct.String())
}

func (ct ConfigType) Equal(o attr.Type) bool {
	other, ok := o.(ConfigType)
	if !ok {
		return false
	}
	return ct == other
}

func (ct ConfigType) String() string {
	return "echostream.ConfigType"
}

func (ct ConfigType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.String
}

func (ct ConfigType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.Type() == nil {
		return diags
	}

	if !in.Type().Is(tftypes.String) {
		err := fmt.Errorf("expected String value, received %T with value: %v", in, in)
		diags.AddAttributeError(
			path,
			"Config Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var config string

	if err := in.As(&config); err != nil {
		diags.AddAttributeError(
			path,
			"Config Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	if config != "" {
		var unknown map[string]any

		if err := json.Unmarshal([]byte(config), &unknown); err != nil {
			diags.AddAttributeError(
				path,
				"Config Type Validation Error",
				"Configs must be a JSON object:\n\n"+err.Error(),
			)
		}
	}

	return diags
}

func (ct ConfigType) ValueFromString(ctx context.Context, in types.String) (types.StringValuable, diag.Diagnostics) {
	if in.IsUnknown() {
		return ConfigUnknown(), nil
	}
	if in.IsNull() {
		return ConfigNull(), nil
	}
	return ConfigValue(in.ValueString()), nil
}

func (ct ConfigType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return ConfigUnknown(), nil
	}
	if in.IsNull() {
		return ConfigNull(), nil
	}
	var s string
	if err := in.As(&s); err != nil {
		return nil, err
	}
	return ConfigValue(s), nil
}

func (ct ConfigType) ValueType(ctx context.Context) attr.Value {
	return Config{}
}

// Config represents a UTF-8 string value.
type Config struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	value string
}

// ConfigNull creates a Config with a null value. Determine whether the value is
// null via the Config type IsNull method.
func ConfigNull() Config {
	return Config{
		state: attr.ValueStateNull,
	}
}

// ConfigUnknown creates a Config with an unknown value. Determine whether the
// value is unknown via the Config type IsUnknown method.
func ConfigUnknown() Config {
	return Config{
		state: attr.ValueStateUnknown,
	}
}

// ConfigValue creates a Config with a known value. Access the value via the Config
// type ValueConfig method.
//
// Setting the deprecated String type Null, Unknown, or Value fields after
// creating a String with this function has no effect.
func ConfigValue(value string) Config {
	return Config{
		state: attr.ValueStateKnown,
		value: value,
	}
}

// Equal returns true if `other` is a Config and has the same value as `c`.
func (c Config) Equal(other attr.Value) bool {
	o, ok := other.(Config)
	if !ok {
		return false
	}
	if c.IsUnknown() != o.IsUnknown() {
		return false
	}
	if c.IsNull() != o.IsNull() {
		return false
	}
	return c.value == o.value
}

// IsNull returns true if the Config represents a null value.
func (c Config) IsNull() bool {
	return c.state == attr.ValueStateNull
}

// IsUnknown returns true if the Config represents a currently unknown value.
func (c Config) IsUnknown() bool {
	return c.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the Config value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (c Config) String() string {
	if c.IsUnknown() {
		return attr.UnknownValueString
	}

	if c.IsNull() {
		return attr.NullValueString
	}

	return fmt.Sprintf("%q", c.value)
}

func (c Config) ToStringValue(ctx context.Context) (types.String, diag.Diagnostics) {
	if c.IsUnknown() {
		return types.StringUnknown(), nil
	}
	if c.IsNull() {
		return types.StringNull(), nil
	}
	return types.StringValue(c.value), nil
}

// ToTerraformValue returns the data contained in the *Config as a tftypes.Value.
func (c Config) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	if c.IsNull() {
		return tftypes.NewValue(tftypes.String, nil), nil
	}
	if c.IsUnknown() {
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), nil
	}
	if err := tftypes.ValidateValue(tftypes.String, c.value); err != nil {
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), err
	}
	return tftypes.NewValue(tftypes.String, c.value), nil
}

// Type returns a ConfigType.
func (c Config) Type(_ context.Context) attr.Type {
	return ConfigType{}
}

// ValueConfig returns the known config value. If Config is null or unknown, returns
// "".
func (c Config) ValueConfig() string {
	return c.value
}
