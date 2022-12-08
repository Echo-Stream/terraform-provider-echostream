package function

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ datasource.DataSourceWithConfigure = &ProcessorFunctionDataSource{}
	_ datasource.DataSourceWithSchema    = &ProcessorFunctionDataSource{}
)

type ProcessorFunctionDataSource struct {
	data *common.ProviderData
}

func (d *ProcessorFunctionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProcessorFunctionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_processor_function"
}

func (d *ProcessorFunctionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config processorFunctionModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, _, diags := readProcessorFunction(ctx, d.data.Client, config.Name.ValueString(), d.data.Tenant); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	} else if data == nil {
		resp.Diagnostics.AddError("ProcessorFunction not found", fmt.Sprintf("Unable to find ProcessorFunction '%s'", config.Name.ValueString()))
		return
	} else {
		config = *data
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *ProcessorFunctionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := dataFunctionAttributes()
	attributes["argument_message_type"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The MessageType passed in to the Function.",
	}
	attributes["return_message_type"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "The MessageType returned by the Function.",
	}
	resp.Schema = schema.Schema{
		Attributes: attributes,
		MarkdownDescription: "[ProcessorFunctions](https://docs.echo.stream/docs/processor-node#processor-function) provide " +
			"reusable message processing and are used in either a ProcessorNode or a CrossTenantSendingNode.",
	}
}
