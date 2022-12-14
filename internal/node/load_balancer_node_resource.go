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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigure   = &LoadBalancerNodeResource{}
	_ resource.ResourceWithImportState = &LoadBalancerNodeResource{}
)

// LoadBalancerNodeResource defines the resource implementation.
type LoadBalancerNodeResource struct {
	data *common.ProviderData
}

type loadBalancerNodeModel struct {
	Description        types.String `tfsdk:"description"`
	Name               types.String `tfsdk:"name"`
	ReceiveMessageType types.String `tfsdk:"receive_message_type"`
	SendMessageType    types.String `tfsdk:"send_message_type"`
}

func (r *LoadBalancerNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LoadBalancerNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}

	if echoResp, err := api.CreateLoadBalancerNode(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		plan.ReceiveMessageType.ValueString(),
		r.data.Tenant,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error creating LoadBalancerNode", err.Error())
		return
	} else {
		if echoResp.CreateLoadBalancerNode.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateLoadBalancerNode.Description)
		} else {
			plan.Description = types.StringNull()
		}
		plan.Name = types.StringValue(echoResp.CreateLoadBalancerNode.Name)
		plan.ReceiveMessageType = types.StringValue(echoResp.CreateLoadBalancerNode.ReceiveMessageType.Name)
		plan.SendMessageType = types.StringValue(echoResp.CreateLoadBalancerNode.SendMessageType.Name)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LoadBalancerNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting LoadBalancerNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *LoadBalancerNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *LoadBalancerNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_node"
}

func (r *LoadBalancerNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading LoadBalancerNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeLoadBalancerNode:
			if node.Description != nil {
				state.Description = types.StringValue(*node.Description)
			} else {
				state.Description = types.StringNull()
			}
			state.Name = types.StringValue(node.Name)
			state.ReceiveMessageType = types.StringValue(node.ReceiveMessageType.Name)
			state.SendMessageType = types.StringValue(node.SendMessageType.Name)
		default:
			resp.Diagnostics.AddError(
				"Expected LoadBalancerNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LoadBalancerNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Node. Must be unique within the Tenant.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
				Validators:          common.NameValidators,
			},
			"receive_message_type": schema.StringAttribute{
				MarkdownDescription: "The MessageType that this Node is capable of receiving.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
			},
			"send_message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that this Node is capable of sending.",
			},
		},
		MarkdownDescription: "[LoadBalancerNodes](https://docs.echo.stream/docs/load-balancer) balance receive messages across all send Edges by " +
			"distributing messages evenly and then randomly distributing any remaining messages. While not required, all Nodes that are targets to a " +
			"LoadBalancerNode's send Edges should be clones of each other. By definition will eliminate guaranteed ordering.",
	}
}

func (r *LoadBalancerNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan loadBalancerNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}

	if echoResp, err := api.UpdateLoadBalancerNode(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		r.data.Tenant,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error updating LoadBalancerNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find LoadBalancerNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.ValueString()))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateLoadBalancerNodeGetNodeLoadBalancerNode:
			if node.Update.Description != nil {
				plan.Description = types.StringValue(*node.Update.Description)
			} else {
				plan.Description = types.StringNull()
			}
			plan.Name = types.StringValue(node.Update.Name)
			plan.ReceiveMessageType = types.StringValue(node.Update.ReceiveMessageType.Name)
			plan.SendMessageType = types.StringValue(node.Update.SendMessageType.Name)
		default:
			resp.Diagnostics.AddError(
				"Expected LoadBalancerNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
