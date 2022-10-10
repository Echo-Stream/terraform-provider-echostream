package message_type

import (
	"context"
	"regexp"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var (
	messageTypeNameValidators []tfsdk.AttributeValidator = []tfsdk.AttributeValidator{
		stringvalidator.LengthBetween(3, 24),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[A-Za-z][A-Za-z0-9\-\_\.]*$`),
			"must contain only lowercase/uppercase alphanumeric characters, \"-\", \"_\", or \".\"",
		),
	}
)

type messageTypeModel struct {
	Auditor           types.String `tfsdk:"auditor"`
	BitmapperTemplate types.String `tfsdk:"bitmapper_template"`
	Description       types.String `tfsdk:"description"`
	InUse             types.Bool   `tfsdk:"in_use"`
	Name              types.String `tfsdk:"name"`
	ProcessorTemplate types.String `tfsdk:"processor_template"`
	Readme            types.String `tfsdk:"readme"`
	Requirements      types.Set    `tfsdk:"requirements"`
	SampleMessage     types.String `tfsdk:"sample_message"`
}

func dataMessageTypeSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"auditor": {
			Computed:            true,
			Description:         "",
			MarkdownDescription: "",
			Type:                types.StringType,
		},
		"bitmapper_template": {
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
			Validators:          messageTypeNameValidators,
		},
		"processor_template": {
			Computed:            true,
			Description:         "",
			MarkdownDescription: "",
			Type:                types.StringType,
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
		"sample_message": {
			Computed:            true,
			Description:         "",
			MarkdownDescription: "",
			Type:                types.StringType,
		},
	}
}

func resourceMessageTypeSchema() map[string]tfsdk.Attribute {
	required := []string{"auditor", "bitmapper_template", "description", "name", "processor_template", "sample_message"}
	schema := dataMessageTypeSchema()
	for key, attribute := range schema {
		if key != "in_use" {
			attribute.Computed = false
		}
		if key == "name" {
			attribute.Validators = append(messageTypeNameValidators, common.NotSystemNameValidator)
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
		}
		if slices.Contains(required, key) {
			attribute.Required = true
		}
		schema[key] = attribute
	}
	return schema
}

func readMessageType(ctx context.Context, client graphql.Client, name string, tenant string) (*messageTypeModel, bool, error) {
	var (
		data     *messageTypeModel
		echoResp *api.ReadMessageTypeResponse
		err      error
		system   bool = false
	)

	if echoResp, err = api.ReadMessageType(ctx, client, name, tenant); err == nil {
		if echoResp.GetMessageType != nil {
			data = &messageTypeModel{}
			data.Auditor = types.String{Value: echoResp.GetMessageType.Auditor}
			data.BitmapperTemplate = types.String{Value: echoResp.GetMessageType.BitmapperTemplate}
			data.Description = types.String{Value: echoResp.GetMessageType.Description}
			data.InUse = types.Bool{Value: echoResp.GetMessageType.InUse}
			data.Name = types.String{Value: echoResp.GetMessageType.Name}
			data.ProcessorTemplate = types.String{Value: echoResp.GetMessageType.ProcessorTemplate}
			if echoResp.GetMessageType.Readme != nil {
				data.Readme = types.String{Value: *echoResp.GetMessageType.Readme}
			} else {
				data.Readme = types.String{Null: true}
			}
			data.Requirements = types.Set{ElemType: types.StringType}
			if len(echoResp.GetMessageType.Requirements) > 0 {
				for _, req := range echoResp.GetMessageType.Requirements {
					data.Requirements.Elems = append(data.Requirements.Elems, types.String{Value: req})
				}
			} else {
				data.Requirements.Null = true
			}
			data.SampleMessage = types.String{Value: echoResp.GetMessageType.SampleMessage}
			if echoResp.GetMessageType.System != nil {
				system = *echoResp.GetMessageType.System
			}
		}
	}

	return data, system, err
}
