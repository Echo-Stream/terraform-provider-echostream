package node

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigure   = &CrossTenantReceivingNodeResource{}
	_ resource.ResourceWithImportState = &CrossTenantReceivingNodeResource{}
)

// ProcessorNodeResource defines the resource implementation.
type CrossTenantReceivingNodeResource struct {
	data *common.ProviderData
}

type crossTenantReceivingNodeModel struct {
	App             types.String `tfsdk:"app"`
	Description     types.String `tfsdk:"description"`
	Name            types.String `tfsdk:"name"`
	SendMessageType types.String `tfsdk:"send_message_type"`
}

func (r *CrossTenantReceivingNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*common.ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.data = data
}

func (r *CrossTenantReceivingNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Cannot create CrossTenantReceivingNode",
		"CrossTenantReceivingNodes are automatically created when their peer CrossTenantSendingNode is created",
	)
}

func (r *CrossTenantReceivingNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state crossTenantReceivingNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting CrossTenantReceivingNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *CrossTenantReceivingNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *CrossTenantReceivingNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cross_tenant_receiving_node"
}

func (r *CrossTenantReceivingNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state crossTenantReceivingNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossTenantReceivingNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeCrossTenantReceivingNode:
			state.App = types.StringValue(node.App.Name)
			if node.Description != nil {
				state.Description = types.StringValue(*node.Description)
			} else {
				state.Description = types.StringNull()
			}
			state.Name = types.StringValue(node.Name)
			state.SendMessageType = types.StringValue(node.SendMessageType.Name)
		default:
			resp.Diagnostics.AddError(
				"Expected CrossTenantReceivingNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CrossTenantReceivingNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The CrossTenantReceivingApp that this Node is associated with.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A human-readable description.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the Node. Automatically generated in the format `<sending_tenant>:<sending_node>`.",
			},
			"send_message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that this Node is capable of sending.",
			},
		},
		MarkdownDescription: "[CrossTenantReceivingNodes](https://docs.echo.stream/docs/cross-tenant-receiving-node) " +
			"receive messages from other Tenants. Created automatically when the other Tenant's CrossTenantSendingApp has " +
			"a CrossTenantSendingNode created in it. This means that you cannot create or update this resource; you may only import " +
			"it and manage it. One per CrossTenantSendingNode.",
	}
}

func (r *CrossTenantReceivingNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Cannot update CrossTenantReceivingNode",
		"CrossTenantReceivingNodes are automatically created when their peer CrossTenantSendingNode is created",
	)
}
