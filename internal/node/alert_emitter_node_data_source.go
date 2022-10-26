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
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSourceWithConfigure = &AlertEmitterNodeDataSource{}

type AlertEmitterNodeDataSource struct {
	data *common.ProviderData
}

func (d *AlertEmitterNodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

type alertEmitterNodeDataSourceModel struct {
	Description     types.String `tfsdk:"description"`
	Name            types.String `tfsdk:"name"`
	SendMessageType types.String `tfsdk:"send_message_type"`
}

func (d *AlertEmitterNodeDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: dataSendNodeSchema(),
		MarkdownDescription: "[AlertEmitterNodes](https://docs.echo.stream/docs/alert-emitter-node) emit alert messages. " +
			"One per Tenant, automatically created when the Tenant is created",
	}, nil
}

func (d *AlertEmitterNodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_emitter_node"
}

func (d *AlertEmitterNodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config alertEmitterNodeDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, d.data.Client, "Alert Emitter", d.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading AlertEmitterNode", err.Error())
		return
	} else if echoResp.GetNode != nil {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeAlertEmitterNode:
			config.Name = types.String{Value: node.Name}
			if node.Description == nil {
				config.Description = types.String{Null: true}
			} else {
				config.Description = types.String{Value: *node.Description}
			}
			config.SendMessageType = types.String{Value: node.SendMessageType.Name}
		default:
			resp.Diagnostics.AddError("Incorrect Node type", fmt.Sprintf("expected AlertEmitterNode, got %v", node.GetTypename()))
			return
		}
	} else {
		resp.Diagnostics.AddWarning("Unable to find node", "'Alert Emitter' node does not exist")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
