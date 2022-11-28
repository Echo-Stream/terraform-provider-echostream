package message_type

import (
	"context"
	"regexp"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	Id                types.String `tfsdk:"id"`
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
			Computed: true,
			MarkdownDescription: "A Python code string that contains a single top-level function definition." +
				" This function must have the signature `(*, message, **kwargs)` where" +
				" message is a string and must return a flat dictionary.",
			Type: types.StringType,
		},
		"bitmapper_template": {
			Computed: true,
			MarkdownDescription: " A Python code string that contains a single top-level function definition." +
				" This function is used as a template when creating custom routing rules in" +
				" RouterNodes that use this MessageType. This function must have the signature" +
				" `(*, context, message, source, **kwargs)` and return an integer.",
			Type: types.StringType,
		},
		"description": {
			Computed:            true,
			MarkdownDescription: "A human-readable description.",
			Type:                types.StringType,
		},
		"id": {
			Computed: true,
			Type:     types.StringType,
		},
		"in_use": {
			Computed:            true,
			MarkdownDescription: "True if this is used by other resources.",
			Type:                types.BoolType,
		},
		"name": {
			MarkdownDescription: "The name of the MessageType.",
			Required:            true,
			Type:                types.StringType,
			Validators:          messageTypeNameValidators,
		},
		"processor_template": {
			Computed: true,
			MarkdownDescription: " A Python code string that contains a single top-leve function definition." +
				" This function is used as a template when creating custom processing in" +
				" ProcessorNodes that use this MessageType. This function must have the signature" +
				" `(*, context, message, source, **kwargs)` and return `None`, a string or a list of strings.",
			Type: types.StringType,
		},
		"readme": {
			Computed:            true,
			MarkdownDescription: "README in MarkDown format.",
			Type:                types.StringType,
		},
		"requirements": {
			Computed:            true,
			MarkdownDescription: "The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format.",
			Type:                types.SetType{ElemType: types.StringType},
			Validators:          []tfsdk.AttributeValidator{common.RequirementsValidator},
		},
		"sample_message": {
			Computed:            true,
			MarkdownDescription: "A sample message.",
			Type:                types.StringType,
		},
	}
}

func resourceMessageTypeSchema() map[string]tfsdk.Attribute {
	required := []string{"auditor", "bitmapper_template", "description", "name", "processor_template", "sample_message"}
	schema := dataMessageTypeSchema()
	for key, attribute := range schema {
		if !slices.Contains([]string{"id", "in_use"}, key) {
			attribute.Computed = false
			if key == "name" {
				attribute.Validators = append(messageTypeNameValidators, common.NotSystemNameValidator)
				attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
			}
			if slices.Contains(required, key) {
				attribute.Required = true
			} else {
				attribute.Optional = true
			}
		}
		schema[key] = attribute
	}
	return schema
}

func readMessageType(ctx context.Context, client graphql.Client, name string, tenant string) (*messageTypeModel, bool, diag.Diagnostics) {
	var (
		data   *messageTypeModel
		diags  diag.Diagnostics
		system bool = false
	)

	if echoResp, err := api.ReadMessageType(ctx, client, name, tenant); err == nil {
		if echoResp.GetMessageType != nil {
			data = &messageTypeModel{}
			data.Auditor = types.StringValue(echoResp.GetMessageType.Auditor)
			data.BitmapperTemplate = types.StringValue(echoResp.GetMessageType.BitmapperTemplate)
			data.Description = types.StringValue(echoResp.GetMessageType.Description)
			data.Id = types.StringValue(echoResp.GetMessageType.Name)
			data.InUse = types.BoolValue(echoResp.GetMessageType.InUse)
			data.Name = types.StringValue(echoResp.GetMessageType.Name)
			data.ProcessorTemplate = types.StringValue(echoResp.GetMessageType.ProcessorTemplate)
			if echoResp.GetMessageType.Readme != nil {
				data.Readme = types.StringValue(*echoResp.GetMessageType.Readme)
			} else {
				data.Readme = types.StringNull()
			}
			if len(echoResp.GetMessageType.Requirements) > 0 {
				var (
					d     diag.Diagnostics
					elems []attr.Value
				)
				for _, req := range echoResp.GetMessageType.Requirements {
					elems = append(elems, types.StringValue(req))
				}
				data.Requirements, d = types.SetValue(types.StringType, elems)
				if d != nil {
					diags = d
				}
			} else {
				data.Requirements = types.SetNull(types.StringType)
			}
			data.SampleMessage = types.StringValue(echoResp.GetMessageType.SampleMessage)
			if echoResp.GetMessageType.System != nil {
				system = *echoResp.GetMessageType.System
			}
		}
	} else {
		diags.AddError("Error reading MessageType", err.Error())
	}

	return data, system, diags
}
