package function

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	ds_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	r_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type functionModel struct {
	Code         types.String `tfsdk:"code"`
	Description  types.String `tfsdk:"description"`
	InUse        types.Bool   `tfsdk:"in_use"`
	Name         types.String `tfsdk:"name"`
	Readme       types.String `tfsdk:"readme"`
	Requirements types.Set    `tfsdk:"requirements"`
}

func dataFunctionAttributes() map[string]ds_schema.Attribute {
	return map[string]ds_schema.Attribute{
		"code": ds_schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The code of the Function in Python string format.",
		},
		"description": ds_schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: " A human-readable description.",
		},
		"in_use": ds_schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "True if this is used by other resources.",
		},
		"name": ds_schema.StringAttribute{
			MarkdownDescription: "The Function name. Must be unique within the Tenant.",
			Required:            true,
			Validators:          common.NameValidators,
		},
		"readme": ds_schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "README in MarkDown format.",
		},
		"requirements": ds_schema.SetAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format.",
		},
	}
}

func resourceFunctionAttributes() map[string]r_schema.Attribute {
	return map[string]r_schema.Attribute{
		"code": r_schema.StringAttribute{
			MarkdownDescription: "The code of the Function in Python string format.",
			Required:            true,
		},
		"description": r_schema.StringAttribute{
			MarkdownDescription: " A human-readable description.",
			Required:            true,
		},
		"in_use": r_schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "True if this is used by other resources.",
		},
		"name": r_schema.StringAttribute{
			MarkdownDescription: "The Function name. Must be unique within the Tenant.",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Required:            true,
			Validators:          append(common.NameValidators, common.NotSystemNameValidator),
		},
		"readme": r_schema.StringAttribute{
			MarkdownDescription: "README in MarkDown format.",
			Optional:            true,
		},
		"requirements": r_schema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format.",
			Optional:            true,
			Validators:          []validator.Set{common.RequirementsValidator},
		},
	}
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
				if function.ReturnMessageType != nil {
					data.ReturnMessageType = types.StringValue(function.ReturnMessageType.Name)
				} else {
					data.ReturnMessageType = types.StringNull()
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
