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
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		Description:         "",
		MarkdownDescription: "",
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

	name := config.App.Value + ":Change Receiver"

	echoResp, err := api.ReadNode(ctx, d.data.Client, name, d.data.Tenant)
	if err != nil {
		resp.Diagnostics.AddError("Error reading AppChangeReceiverNode", err.Error())
		return
	}
	if echoResp.GetNode != nil {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeAppChangeReceiverNode:
			config.App = types.String{Value: node.App.GetName()}
			config.Name = types.String{Value: node.Name}
			if node.Description == nil {
				config.Description = types.String{Null: true}
			} else {
				config.Description = types.String{Value: *node.Description}
			}
			config.ReceiveMessageType = types.String{Value: node.ReceiveMessageType.Name}
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
