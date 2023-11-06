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
var _ datasource.DataSourceWithConfigure = &ExternalNodeDataSource{}

type ExternalNodeDataSource struct {
	data *common.ProviderData
}

func (d *ExternalNodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ExternalNodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_node"
}

func (d *ExternalNodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config externalNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, d.data.Client, config.Name.ValueString(), d.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ExternalNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeExternalNode:
			switch app := (node.App).(type) {
			case *api.ExternalNodeFieldsAppCrossAccountApp:
				config.App = types.StringValue(app.Name)
			case *api.ExternalNodeFieldsAppExternalApp:
				config.App = types.StringValue(app.Name)
			default:
				resp.Diagnostics.AddError(
					"Invalid App type",
					fmt.Sprintf("Expected CrossAccountApp or ExternalApp, got %s", *app.GetTypename()),
				)
			}
			if node.Config != nil {
				config.Config = common.ConfigValue(*node.Config)
			} else {
				config.Config = common.ConfigNull()
			}
			if node.Description != nil {
				config.Description = types.StringValue(*node.Description)
			} else {
				config.Description = types.StringNull()
			}
			config.Name = types.StringValue(node.Name)
			if node.ReceiveMessageType != nil {
				config.ReceiveMessageType = types.StringValue(node.ReceiveMessageType.Name)
			} else {
				config.ReceiveMessageType = types.StringNull()
			}
			if node.SendMessageType != nil {
				config.SendMessageType = types.StringValue(node.SendMessageType.Name)
			} else {
				config.SendMessageType = types.StringNull()
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ExternalNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), config.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *ExternalNodeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ExternalApp or CrossAccountApp this Node is associated with.",
			},
			"config": schema.StringAttribute{
				Computed:            true,
				CustomType:          common.ConfigType{},
				MarkdownDescription: "The config, in JSON object format (i.e. - dict, map).",
				Sensitive:           true,
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A human-readable description.",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Node. Must be unique within the Tenant.",
				Required:            true,
			},
			"receive_message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that this Node is capable of receiving.",
			},
			"send_message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that this Node is capable of sending.",
			},
		},
		MarkdownDescription: "[ExternalNodes](https://docs.echo.stream/docs/external-node) exist outside the " +
			"EchoStream Cloud. Can be part of an ExternalApp or CrossAccountApp. You may use any computing resource " +
			"or language that you want to implement them.",
	}
}
