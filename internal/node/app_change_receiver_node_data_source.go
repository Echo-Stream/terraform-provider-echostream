package node

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSourceWithConfigure = &AppChangeReceiverNodeDataSource{}

type AppChangeReceiverNodeDataSource struct {
	data *common.ProviderData
}

func (d *AppChangeReceiverNodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

type appChangeReceiverNodeDataSourceModel struct {
	App                types.String `tfsdk:"app"`
	Description        types.String `tfsdk:"description"`
	Name               types.String `tfsdk:"name"`
	ReceiveMessageType types.String `tfsdk:"receive_message_type"`
}

func (d *AppChangeReceiverNodeDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := dataReceiveNodeSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"app": {
				MarkdownDescription: "The App for this AppChangeReceiverNode",
				Required:            true,
				Type:                types.StringType,
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		MarkdownDescription: "AppChangeReceiverNodes receive change messages from the AppChangeRouterNode. One per App, created when the App is created.",
	}, nil
}

func (d *AppChangeReceiverNodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_change_receiver_node"
}

func (d *AppChangeReceiverNodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config appChangeReceiverNodeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := config.App.ValueString() + ":Change Receiver"

	if echoResp, err := api.ReadNode(ctx, d.data.Client, name, d.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading AppChangeReceiverNode", err.Error())
		return
	} else if echoResp.GetNode != nil {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeAppChangeReceiverNode:
			config.App = types.StringValue(node.App.GetName())
			config.Name = types.StringValue(node.Name)
			if node.Description == nil {
				config.Description = types.StringNull()
			} else {
				config.Description = types.StringValue(*node.Description)
			}
			config.ReceiveMessageType = types.StringValue(node.ReceiveMessageType.Name)
		default:
			resp.Diagnostics.AddError("Incorrect Node type", fmt.Sprintf("expected AppChangeReceiverNode, got %v", node.GetTypename()))
			return
		}
	} else {
		resp.Diagnostics.AddWarning("Unable to find node", fmt.Sprintf("'%s' node does not exist", name))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
