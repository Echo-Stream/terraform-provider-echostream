package node

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
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
		description = &plan.Description.Value
	}

	if echoResp, err := api.CreateLoadBalancerNode(
		ctx,
		r.data.Client,
		plan.Name.Value,
		plan.ReceiveMessageType.Value,
		r.data.Tenant,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error creating LoadBalancerNode", err.Error())
		return
	} else {
		if echoResp.CreateLoadBalancerNode.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateLoadBalancerNode.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		plan.Name = types.String{Value: echoResp.CreateLoadBalancerNode.Name}
		plan.ReceiveMessageType = types.String{Value: echoResp.CreateLoadBalancerNode.ReceiveMessageType.Name}
		plan.SendMessageType = types.String{Value: echoResp.CreateLoadBalancerNode.SendMessageType.Name}
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

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting LoadBalancerNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *LoadBalancerNodeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := dataSendReceiveNodeSchema()
	for key, attribute := range schema {
		switch key {
		case "description":
			attribute.Computed = false
			attribute.Optional = true
		case "name":
			attribute.Computed = false
			attribute.Required = true
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
			attribute.Validators = common.NameValidators
		case "receive_message_type":
			attribute.Computed = false
			attribute.Required = true
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
		}
		schema[key] = attribute
	}
	return tfsdk.Schema{
		Attributes:          schema,
		Description:         "LoadBalancerNodes allow for the distribution of messages to a group of like Nodes",
		MarkdownDescription: "LoadBalancerNodes allow for the distribution of messages to a group of like Nodes",
	}, nil
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

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading LoadBalancerNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeLoadBalancerNode:
			if node.Description != nil {
				state.Description = types.String{Value: *node.Description}
			} else {
				state.Description = types.String{Null: true}
			}
			state.Name = types.String{Value: node.Name}
			state.ReceiveMessageType = types.String{Value: node.ReceiveMessageType.Name}
			state.SendMessageType = types.String{Value: node.SendMessageType.Name}
		default:
			resp.Diagnostics.AddError(
				"Expected LoadBalancerNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
		description = &plan.Description.Value
	}

	if echoResp, err := api.UpdateLoadBalancerNode(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error updating LoadBalancerNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find LoadBalancerNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.Value))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateLoadBalancerNodeGetNodeLoadBalancerNode:
			if node.Update.Description != nil {
				plan.Description = types.String{Value: *node.Update.Description}
			} else {
				plan.Description = types.String{Null: true}
			}
			plan.Name = types.String{Value: node.Update.Name}
			plan.ReceiveMessageType = types.String{Value: node.Update.ReceiveMessageType.Name}
			plan.SendMessageType = types.String{Value: node.Update.SendMessageType.Name}
		default:
			resp.Diagnostics.AddError(
				"Expected LoadBalancerNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
