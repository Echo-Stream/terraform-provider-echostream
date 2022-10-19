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
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithImportState = &CrossTenantReceivingNodeResource{}
	_ resource.ResourceWithModifyPlan  = &CrossTenantReceivingNodeResource{}
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
	var plan crossTenantReceivingNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, plan.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossTenantReceivingNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeCrossTenantReceivingNode:
			plan.App = types.String{Value: node.App.Name}
			if node.Description != nil {
				plan.Description = types.String{Value: *node.Description}
			} else {
				plan.Description = types.String{Null: true}
			}
			plan.Name = types.String{Value: node.Name}
			plan.SendMessageType = types.String{Value: node.SendMessageType.Name}
		default:
			resp.Diagnostics.AddError(
				"Expected CrossTenantReceivingNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CrossTenantReceivingNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state crossTenantReceivingNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting CrossTenantReceivingNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *CrossTenantReceivingNodeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := dataSendNodeSchema()
	name := schema["name"]
	name.Computed = false
	name.Required = true
	name.Validators = common.NameValidators
	schema["name"] = name
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"app": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		Description:         "CrossTenantReceivingNodes receive messages from other Tenants",
		MarkdownDescription: "CrossTenantReceivingNodes areceive messages from other Tenants",
	}, nil
}

func (r *CrossTenantReceivingNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *CrossTenantReceivingNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cross_tenant_receiving_node"
}

func (r *CrossTenantReceivingNodeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state crossTenantReceivingNodeModel

	// If the entire state is null, resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// If the entire plan is null, the resource is planned for destruction.
	if req.Plan.Raw.IsNull() {
		resp.Diagnostics.AddWarning(
			"Deleting CrossTenantReceivingNode",
			"This will also delete the sending Tenant's CrossTenantSendingNode!",
		)
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var plan crossTenantReceivingNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if !plan.Name.Equal(state.Name) {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Cannot change a CrossTenantReceivingNode's `name`",
			"This would break the connection with the sending Tenant's CrossTenantSendingNode.",
		)
	}
}

func (r *CrossTenantReceivingNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state crossTenantReceivingNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossTenantReceivingNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeCrossTenantReceivingNode:
			state.App = types.String{Value: node.App.Name}
			if node.Description != nil {
				state.Description = types.String{Value: *node.Description}
			} else {
				state.Description = types.String{Null: true}
			}
			state.Name = types.String{Value: node.Name}
			state.SendMessageType = types.String{Value: node.SendMessageType.Name}
		default:
			resp.Diagnostics.AddError(
				"Expected CrossTenantReceivingNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CrossTenantReceivingNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan crossTenantReceivingNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string

	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if echoResp, err := api.UpdateCrossTenantReceivingNode(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error updating CrossTenantReceivingNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find CrossTenantReceivingNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.Value))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateCrossTenantReceivingNodeGetNodeCrossTenantReceivingNode:
			plan.App = types.String{Value: node.Update.App.Name}
			if node.Update.Description != nil {
				plan.Description = types.String{Value: *node.Update.Description}
			} else {
				plan.Description = types.String{Null: true}
			}
			plan.Name = types.String{Value: node.Update.Name}
			plan.SendMessageType = types.String{Value: node.Update.SendMessageType.Name}
		default:
			resp.Diagnostics.AddError(
				"Expected CrossTenantReceivingNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
