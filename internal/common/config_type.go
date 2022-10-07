package common

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ xattr.TypeWithValidate = ConfigType{}
	_ attr.Value             = Config{}
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

func (ct ConfigType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return Config{Unknown: true}, nil
	}
	if in.IsNull() {
		return Config{Null: true}, nil
	}
	var s string
	if err := in.As(&s); err != nil {
		return nil, err
	}
	return Config{Value: s}, nil
}

func (ct ConfigType) ValueType(ctx context.Context) attr.Value {
	return Config{}
}

// Config represents a UTF-8 string value.
type Config struct {
	// Unknown will be true if the value is not yet known.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	Value string
}

// Equal returns true if `other` is a Config and has the same value as `c`.
func (c Config) Equal(other attr.Value) bool {
	o, ok := other.(Config)
	if !ok {
		return false
	}
	if c.Unknown != o.Unknown {
		return false
	}
	if c.Null != o.Null {
		return false
	}
	return c.Value == o.Value
}

// IsNull returns true if the Config represents a null value.
func (c Config) IsNull() bool {
	return c.Null
}

// IsUnknown returns true if the Config represents a currently unknown value.
func (c Config) IsUnknown() bool {
	return c.Unknown
}

// String returns a human-readable representation of the Config value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (c Config) String() string {
	if c.Unknown {
		return attr.UnknownValueString
	}

	if c.Null {
		return attr.NullValueString
	}

	return fmt.Sprintf("%q", c.Value)
}

// ToTerraformValue returns the data contained in the *Config as a tftypes.Value.
func (c Config) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	if c.Null {
		return tftypes.NewValue(tftypes.String, nil), nil
	}
	if c.Unknown {
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), nil
	}
	if err := tftypes.ValidateValue(tftypes.String, c.Value); err != nil {
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), err
	}
	return tftypes.NewValue(tftypes.String, c.Value), nil
}

// Type returns a ConfigType.
func (c Config) Type(_ context.Context) attr.Type {
	return ConfigType{}
}
