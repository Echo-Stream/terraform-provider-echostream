package function

import (
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
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
