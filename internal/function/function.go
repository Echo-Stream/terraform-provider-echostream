package function

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type functionModel struct {
	Code         types.String `tfsdk:"code"`
	Description  types.String `tfsdk:"description"`
	InUse        types.Bool   `tfsdk:"in_use"`
	Name         types.String `tfsdk:"name"`
	Readme       types.String `tfsdk:"readme"`
	Requirements types.Set    `tfsdk:"requirements"`
}

func dataFunctionSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"code": {
			Computed:            true,
			MarkdownDescription: "The code of the Function in Python string format.",
			Type:                types.StringType,
		},
		"description": {
			Computed:            true,
			MarkdownDescription: " A human-readable description.",
			Type:                types.StringType,
		},
		"in_use": {
			Computed:            true,
			MarkdownDescription: "True if this is used by other resources.",
			Type:                types.BoolType,
		},
		"name": {
			MarkdownDescription: "The Function name. Must be unique within the Tenant.",
			Required:            true,
			Type:                types.StringType,
			Validators:          common.NameValidators,
		},
		"readme": {
			Computed:            true,
			MarkdownDescription: "README in MarkDown format.",
			Optional:            true,
			Type:                types.StringType,
		},
		"requirements": {
			Computed:            true,
			MarkdownDescription: "The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format.",
			Optional:            true,
			Type:                types.SetType{ElemType: types.StringType},
			Validators:          []tfsdk.AttributeValidator{common.RequirementsValidator},
		},
	}
}

func resourceFunctionSchema() map[string]tfsdk.Attribute {
	required := []string{"code", "description", "name"}
	schema := dataFunctionSchema()
	for key, attribute := range schema {
		if key != "in_use" {
			attribute.Computed = false
		}
		if key == "name" {
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
			attribute.Validators = append(common.NameValidators, common.NotSystemNameValidator)
		}
		if slices.Contains(required, key) {
			attribute.Required = true
		}
		schema[key] = attribute
	}
	return schema
}

type bitmapperFunctionModel struct {
	ArgumentMessageType types.String `tfsdk:"argument_message_type"`
	Code                types.String `tfsdk:"code"`
	Description         types.String `tfsdk:"description"`
	InUse               types.Bool   `tfsdk:"in_use"`
	Name                types.String `tfsdk:"name"`
	Readme              types.String `tfsdk:"readme"`
	Requirements        types.Set    `tfsdk:"requirements"`
}

func dataBitmapperFunctionSchema() map[string]tfsdk.Attribute {
	schema := dataFunctionSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"argument_message_type": {
				Computed:            true,
				MarkdownDescription: "The MessageType passed in to the Function.",
				Type:                types.StringType,
			},
		},
	)
	return schema
}

func resourceBitmapperFunctionSchema() map[string]tfsdk.Attribute {
	schema := resourceFunctionSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"argument_message_type": {
				MarkdownDescription: "The MessageType passed in to the Function.",
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Required:            true,
				Type:                types.StringType,
			},
		},
	)
	return schema
}

type processorFunctionModel struct {
	ArgumentMessageType types.String `tfsdk:"argument_message_type"`
	Code                types.String `tfsdk:"code"`
	Description         types.String `tfsdk:"description"`
	InUse               types.Bool   `tfsdk:"in_use"`
	Name                types.String `tfsdk:"name"`
	Readme              types.String `tfsdk:"readme"`
	Requirements        types.Set    `tfsdk:"requirements"`
	ReturnMessageType   types.String `tfsdk:"return_message_type"`
}

func dataProcessorFunctionSchema() map[string]tfsdk.Attribute {
	schema := dataFunctionSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"argument_message_type": {
				Computed:            true,
				MarkdownDescription: "The MessageType passed in to the Function.",
				Type:                types.StringType,
			},
			"return_message_type": {
				Computed:            true,
				MarkdownDescription: "The MessageType returned by the Function.",
				Type:                types.StringType,
			},
		},
	)
	return schema
}

func resourceProcessorFunctionSchema() map[string]tfsdk.Attribute {
	schema := resourceFunctionSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"argument_message_type": {
				MarkdownDescription: "The MessageType passed in to the Function.",
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Required:            true,
				Type:                types.StringType,
			},
			"return_message_type": {
				MarkdownDescription: "The MessageType returned by the Function.",
				Optional:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Type:                types.StringType,
			},
		},
	)
	return schema
}

func readApiAuthenicatorFunction(ctx context.Context, client graphql.Client, name string, tenant string) (*functionModel, bool, diag.Diagnostics) {
	var (
		data   *functionModel
		diags  diag.Diagnostics
		system bool = false
	)

	if echoResp, err := api.ReadFunction(ctx, client, name, tenant); err == nil {
		if echoResp.GetFunction != nil {
			switch function := (*echoResp.GetFunction).(type) {
			case *api.ReadFunctionGetFunctionApiAuthenticatorFunction:
				data = &functionModel{}
				data.Code = types.StringValue(function.Code)
				data.Description = types.StringValue(function.Description)
				data.InUse = types.BoolValue(function.InUse)
				data.Name = types.StringValue(function.Name)
				if function.Readme != nil {
					data.Readme = types.StringValue(*function.Readme)
				} else {
					data.Readme = types.StringNull()
				}
				if len(function.Requirements) > 0 {
					var (
						d     diag.Diagnostics
						elems []attr.Value
					)
					for _, req := range function.Requirements {
						elems = append(elems, types.StringValue(req))
					}
					data.Requirements, d = types.SetValue(types.StringType, elems)
					if d != nil {
						diags = d
					}
				} else {
					data.Requirements = types.SetNull(types.StringType)
				}
				if function.System != nil {
					system = *function.System
				}
			default:
				diags.AddError("Invalid Function type", fmt.Sprintf("'%s' is incorrect Function type", data.Name.String()))
			}
		}
	} else {
		diags.AddError("Error reading ApiAuthenticatorFunction", err.Error())
	}

	return data, system, diags
}

func readBitmapperFunction(ctx context.Context, client graphql.Client, name string, tenant string) (*bitmapperFunctionModel, bool, diag.Diagnostics) {
	var (
		data   *bitmapperFunctionModel
		diags  diag.Diagnostics
		system bool = false
	)

	if echoResp, err := api.ReadFunction(ctx, client, name, tenant); err == nil {
		if echoResp.GetFunction != nil {
			switch function := (*echoResp.GetFunction).(type) {
			case *api.ReadFunctionGetFunctionBitmapperFunction:
				data = &bitmapperFunctionModel{}
				data.ArgumentMessageType = types.StringValue(function.ArgumentMessageType.Name)
				data.Code = types.StringValue(function.Code)
				data.Description = types.StringValue(function.Description)
				data.InUse = types.BoolValue(function.InUse)
				data.Name = types.StringValue(function.Name)
				if function.Readme != nil {
					data.Readme = types.StringValue(*function.Readme)
				} else {
					data.Readme = types.StringNull()
				}
				if len(function.Requirements) > 0 {
					var (
						d     diag.Diagnostics
						elems []attr.Value
					)
					for _, req := range function.Requirements {
						elems = append(elems, types.StringValue(req))
					}
					data.Requirements, d = types.SetValue(types.StringType, elems)
					if d != nil {
						diags = d
					}
				} else {
					data.Requirements = types.SetNull(types.StringType)
				}
				if function.System != nil {
					system = *function.System
				}
			default:
				diags.AddError("Invalid Function type", fmt.Sprintf("'%s' is incorrect Function type", data.Name.String()))
			}
		}
	} else {
		diags.AddError("Error reading ApiAuthenticatorFunction", err.Error())
	}

	return data, system, diags
}

func readProcessorFunction(ctx context.Context, client graphql.Client, name string, tenant string) (*processorFunctionModel, bool, diag.Diagnostics) {
	var (
		data   *processorFunctionModel
		diags  diag.Diagnostics
		system bool = false
	)

	if echoResp, err := api.ReadFunction(ctx, client, name, tenant); err == nil {
		if echoResp.GetFunction != nil {
			switch function := (*echoResp.GetFunction).(type) {
			case *api.ReadFunctionGetFunctionProcessorFunction:
				data = &processorFunctionModel{}
				data.ArgumentMessageType = types.StringValue(function.ArgumentMessageType.Name)
				data.Code = types.StringValue(function.Code)
				data.Description = types.StringValue(function.Description)
				data.InUse = types.BoolValue(function.InUse)
				data.Name = types.StringValue(function.Name)
				if function.Readme != nil {
					data.Readme = types.StringValue(*function.Readme)
				} else {
					data.Readme = types.StringNull()
				}
				if len(function.Requirements) > 0 {
					var (
						d     diag.Diagnostics
						elems []attr.Value
					)
					for _, req := range function.Requirements {
						elems = append(elems, types.StringValue(req))
					}
					data.Requirements, d = types.SetValue(types.StringType, elems)
					if d != nil {
						diags = d
					}
				} else {
					data.Requirements = types.SetNull(types.StringType)
				}
				data.ReturnMessageType = types.StringValue(function.ReturnMessageType.Name)
				if function.System != nil {
					system = *function.System
				}
			default:
				diags.AddError("Invalid Function type", fmt.Sprintf("'%s' is incorrect Function type", data.Name.String()))
			}
		}
	} else {
		diags.AddError("Error reading ApiAuthenticatorFunction", err.Error())
	}

	return data, system, diags
}
