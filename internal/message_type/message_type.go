package message_type

import (
	"context"
	"regexp"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var (
	messageTypeNameValidators []validator.String = []validator.String{
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
