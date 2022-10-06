package function

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/Khan/genqlient/graphql"
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
			Description:         "",
			MarkdownDescription: "",
			Type:                types.StringType,
		},
		"description": {
			Computed:            true,
			Description:         "",
			MarkdownDescription: "",
			Type:                types.StringType,
		},
		"in_use": {
			Computed:            true,
			Description:         "",
			MarkdownDescription: "",
			Type:                types.BoolType,
		},
		"name": {
			Description:         "",
			MarkdownDescription: "",
			Required:            true,
			Type:                types.StringType,
			Validators:          common.NameValidators,
		},
		"readme": {
			Computed:            true,
			Description:         "",
			MarkdownDescription: "",
			Optional:            true,
			Type:                types.StringType,
		},
		"requirements": {
			Computed:            true,
			Description:         "",
			MarkdownDescription: "",
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
			attribute.Validators = append(common.NameValidators, common.NotSystemNameValidator)
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
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
				Description:         "",
				MarkdownDescription: "",
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
				Description:         "",
				MarkdownDescription: "",
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
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
			"return_message_type": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
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
				Description:         "",
				MarkdownDescription: "",
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Required:            true,
				Type:                types.StringType,
			},
			"return_message_type": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Type:                types.StringType,
			},
		},
	)
	return schema
}

func readFunction(ctx context.Context, client graphql.Client, name string, tenant string, data *functionModel) (bool, error) {
	var (
		echoResp *api.ReadFunctionResponse
		err      error
		system   bool = false
	)

	if echoResp, err = api.ReadFunction(ctx, client, name, tenant); err == nil {
		if echoResp.GetFunction != nil {
			switch function := (*echoResp.GetFunction).(type) {
			case *api.ReadFunctionGetFunctionApiAuthenticatorFunction:
				data.Code = types.String{Value: function.Code}
				data.Description = types.String{Value: function.Description}
				data.InUse = types.Bool{Value: function.InUse}
				data.Name = types.String{Value: function.Name}
				if function.Readme != nil {
					data.Readme = types.String{Value: *function.Readme}
				} else {
					data.Readme = types.String{Null: true}
				}
				if len(function.Requirements) > 0 {
					data.Requirements = types.Set{ElemType: types.StringType}
					for _, req := range function.Requirements {
						data.Requirements.Elems = append(data.Requirements.Elems, types.String{Value: req})
					}
				} else {
					data.Requirements.Null = true
				}
				if function.System != nil {
					system = *function.System
				}
			default:
				err = fmt.Errorf("'%s' is incorrect Function type", data.Name.String())
			}
		} else {
			err = fmt.Errorf("'%s' Function does not exist", data.Name.String())
		}
	}

	return system, err
}

func readBitmapperFunction(ctx context.Context, client graphql.Client, name string, tenant string, data *bitmapperFunctionModel) (bool, error) {
	var (
		echoResp *api.ReadFunctionResponse
		err      error
		system   bool = false
	)

	if echoResp, err = api.ReadFunction(ctx, client, name, tenant); err == nil {
		if echoResp.GetFunction != nil {
			switch function := (*echoResp.GetFunction).(type) {
			case *api.ReadFunctionGetFunctionBitmapperFunction:
				data.ArgumentMessageType = types.String{Value: function.ArgumentMessageType.Name}
				data.Code = types.String{Value: function.Code}
				data.Description = types.String{Value: function.Description}
				data.InUse = types.Bool{Value: function.InUse}
				data.Name = types.String{Value: function.Name}
				if function.Readme != nil {
					data.Readme = types.String{Value: *function.Readme}
				} else {
					data.Readme = types.String{Null: true}
				}
				if len(function.Requirements) > 0 {
					data.Requirements = types.Set{ElemType: types.StringType}
					for _, req := range function.Requirements {
						data.Requirements.Elems = append(data.Requirements.Elems, types.String{Value: req})
					}
				} else {
					data.Requirements.Null = true
				}
				if function.System != nil {
					system = *function.System
				}
			default:
				err = fmt.Errorf("'%s' is incorrect Function type", data.Name.String())
			}
		} else {
			err = fmt.Errorf("'%s' Function does not exist", data.Name.String())
		}
	}

	return system, err
}

func readProcessorFunction(ctx context.Context, client graphql.Client, name string, tenant string, data *processorFunctionModel) (bool, error) {
	var (
		echoResp *api.ReadFunctionResponse
		err      error
		system   bool = false
	)

	if echoResp, err = api.ReadFunction(ctx, client, name, tenant); err == nil {
		if echoResp.GetFunction != nil {
			switch function := (*echoResp.GetFunction).(type) {
			case *api.ReadFunctionGetFunctionProcessorFunction:
				data.ArgumentMessageType = types.String{Value: function.ArgumentMessageType.Name}
				data.Code = types.String{Value: function.Code}
				data.Description = types.String{Value: function.Description}
				data.InUse = types.Bool{Value: function.InUse}
				data.Name = types.String{Value: function.Name}
				if function.Readme != nil {
					data.Readme = types.String{Value: *function.Readme}
				} else {
					data.Readme = types.String{Null: true}
				}
				if len(function.Requirements) > 0 {
					data.Requirements = types.Set{ElemType: types.StringType}
					for _, req := range function.Requirements {
						data.Requirements.Elems = append(data.Requirements.Elems, types.String{Value: req})
					}
				} else {
					data.Requirements.Null = true
				}
				data.ReturnMessageType = types.String{Value: function.ReturnMessageType.Name}
				if function.System != nil {
					system = *function.System
				}
			default:
				err = fmt.Errorf("'%s' is incorrect Function type", data.Name.String())
			}
		} else {
			err = fmt.Errorf("'%s' Function does not exist", data.Name.String())
		}
	}

	return system, err
}
