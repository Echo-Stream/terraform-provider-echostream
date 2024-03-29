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
var _ datasource.DataSourceWithConfigure = &AppChangeRouterNodeDataSource{}

type AppChangeRouterNodeDataSource struct {
	data *common.ProviderData
}

func (d *AppChangeRouterNodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

type appChangeRouterNodeDataSourceModel struct {
	Description        types.String `tfsdk:"description"`
	Name               types.String `tfsdk:"name"`
	ReceiveMessageType types.String `tfsdk:"receive_message_type"`
	SendMessageType    types.String `tfsdk:"send_message_type"`
}

func (d *AppChangeRouterNodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_change_router_node"
}

func (d *AppChangeRouterNodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config appChangeRouterNodeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, d.data.Client, "App Change Router", d.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading AppChangeRouterNode", err.Error())
		return
	} else if echoResp.GetNode != nil {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeAppChangeRouterNode:
			config.Name = types.StringValue(node.Name)
			if node.Description == nil {
				config.Description = types.StringNull()
			} else {
				config.Description = types.StringValue(*node.Description)
			}
			config.ReceiveMessageType = types.StringValue(node.ReceiveMessageType.Name)
			config.SendMessageType = types.StringValue(node.SendMessageType.Name)
		default:
			resp.Diagnostics.AddError("Incorrect Node type", fmt.Sprintf("expected AppChangeRouterNode, got %v", node.GetTypename()))
			return
		}
	} else {
		resp.Diagnostics.AddWarning("Unable to find node", "'App Change Router' node does not exist")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *AppChangeRouterNodeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: sendReceiveNodeDataSourceAttributes(),
		MarkdownDescription: "[AppChangeRouterNodes](https://docs.echo.stream/docs/app-change-router-node) route change messages " +
			"to the appropriate App. One per Tenant, automatically created when the Tenant is created.",
	}
}
