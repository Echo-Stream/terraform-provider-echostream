package message_type

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSourceWithConfigure = &MessageTypeDataSource{}

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

func (d *MessageTypeDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:          dataMessageTypeSchema(),
		Description:         "",
		MarkdownDescription: "",
	}, nil
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

	if err := readMessageType(ctx, d.data.Client, config.Name.Value, d.data.Tenant, true, &config); err != nil {
		resp.Diagnostics.AddError("Error reading MessageType", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
