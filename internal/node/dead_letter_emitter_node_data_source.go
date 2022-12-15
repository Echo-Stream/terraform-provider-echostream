package node

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSourceWithConfigure = &DeadLetterEmitterNodeDataSource{}

type DeadLetterEmitterNodeDataSource struct {
	data *common.ProviderData
}

func (d *DeadLetterEmitterNodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

type deadLetterEmitterNodeDataSourceModel struct {
	Description     types.String `tfsdk:"description"`
	Name            types.String `tfsdk:"name"`
	SendMessageType types.String `tfsdk:"send_message_type"`
}

func (d *DeadLetterEmitterNodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dead_letter_emitter_node"
}

func (d *DeadLetterEmitterNodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config deadLetterEmitterNodeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, d.data.Client, "Dead Letter Emitter", d.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading DeadLetterEmitterNode", err.Error())
		return
	} else if echoResp.GetNode != nil {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeDeadLetterEmitterNode:
			config.Name = types.StringValue(node.Name)
			if node.Description == nil {
				config.Description = types.StringNull()
			} else {
				config.Description = types.StringValue(*node.Description)
			}
			config.SendMessageType = types.StringValue(node.SendMessageType.Name)
		default:
			resp.Diagnostics.AddError("Incorrect Node type", fmt.Sprintf("expected DeadLetterEmitterNode, got %v", node.GetTypename()))
			return
		}
	} else {
		resp.Diagnostics.AddWarning("Unable to find node", "'Dead Letter Emitter' node does not exist")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *DeadLetterEmitterNodeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: dataSendNodeAttributes(),
		MarkdownDescription: "[DeadLetterEmitterNodes](https://docs.echo.stream/docs/dead-letter-emitter-node) emit dead letters (i.e. - " +
			"undeliverable messages). One per Tenant, automatically created when the Tenant is created.",
	}
}
