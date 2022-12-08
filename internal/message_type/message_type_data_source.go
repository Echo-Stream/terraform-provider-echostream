package message_type

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ datasource.DataSourceWithConfigure = &MessageTypeDataSource{}
	_ datasource.DataSourceWithSchema    = &MessageTypeDataSource{}
)

type MessageTypeDataSource struct {
	data *common.ProviderData
}

func (d *MessageTypeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*common.ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.data = data
}

func (d *MessageTypeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_message_type"
}

func (d *MessageTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config messageTypeModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, _, diags := readMessageType(ctx, d.data.Client, config.Name.ValueString(), d.data.Tenant); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	} else if data == nil {
		resp.Diagnostics.AddError("MessageType not found", fmt.Sprintf("Unable to find MessageType '%s'", config.Name.ValueString()))
		return
	} else {
		config = *data
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *MessageTypeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"auditor": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "A Python code string that contains a single top-level function definition." +
					" This function must have the signature `(*, message, **kwargs)` where" +
					" message is a string and must return a flat dictionary.",
			},
			"bitmapper_template": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: " A Python code string that contains a single top-level function definition." +
					" This function is used as a template when creating custom routing rules in" +
					" RouterNodes that use this MessageType. This function must have the signature" +
					" `(*, context, message, source, **kwargs)` and return an integer.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A human-readable description.",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"in_use": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "True if this is used by other resources.",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the MessageType.",
				Required:            true,
				Validators:          messageTypeNameValidators,
			},
			"processor_template": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: " A Python code string that contains a single top-leve function definition." +
					" This function is used as a template when creating custom processing in" +
					" ProcessorNodes that use this MessageType. This function must have the signature" +
					" `(*, context, message, source, **kwargs)` and return `None`, a string or a list of strings.",
			},
			"readme": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "README in MarkDown format.",
			},
			"requirements": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format.",
			},
			"sample_message": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A sample message.",
			},
		},
		MarkdownDescription: "A specific [MessageType](https://docs.echo.stream/docs/message-types) in the Tenant. " +
			"All messages sent or received must be loosely associated (via Node and Edge typing) with a MessageType.",
	}
}
