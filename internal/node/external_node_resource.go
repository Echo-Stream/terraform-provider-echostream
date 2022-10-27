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
var _ resource.ResourceWithImportState = &ExternalNodeResource{}

// ExternalNodeResource defines the resource implementation.
type ExternalNodeResource struct {
	data *common.ProviderData
}

type externalNodeModel struct {
	App                types.String  `tfsdk:"app"`
	Config             common.Config `tfsdk:"config"`
	Description        types.String  `tfsdk:"description"`
	Name               types.String  `tfsdk:"name"`
	ReceiveMessageType types.String  `tfsdk:"receive_message_type"`
	SendMessageType    types.String  `tfsdk:"send_message_type"`
}

func (r *ExternalNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ExternalNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan externalNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config             *string
		description        *string
		receiveMessageType *string
		sendMessageType    *string
	)

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.ReceiveMessageType.IsNull() || plan.ReceiveMessageType.IsUnknown()) {
		receiveMessageType = &plan.ReceiveMessageType.Value
	}
	if !(plan.SendMessageType.IsNull() || plan.SendMessageType.IsUnknown()) {
		sendMessageType = &plan.SendMessageType.Value
	}

	if echoResp, err := api.CreateExternalNode(
		ctx,
		r.data.Client,
		plan.App.Value,
		plan.Name.Value,
		r.data.Tenant,
		config,
		description,
		receiveMessageType,
		sendMessageType,
	); err != nil {
		resp.Diagnostics.AddError("Error creating ExternalNode", err.Error())
		return
	} else {
		switch app := (echoResp.CreateExternalNode.App).(type) {
		case *api.ExternalNodeFieldsAppCrossAccountApp:
			plan.App = types.String{Value: app.Name}
		case *api.ExternalNodeFieldsAppExternalApp:
			plan.App = types.String{Value: app.Name}
		default:
			resp.Diagnostics.AddError(
				"Invalid App type",
				fmt.Sprintf("Expected CrossAccountApp or ExternalApp, got %s", *app.GetTypename()),
			)
		}
		if echoResp.CreateExternalNode.Config != nil {
			plan.Config = common.Config{Value: *echoResp.CreateExternalNode.Config}
		} else {
			plan.Config = common.Config{Null: true}
		}
		if echoResp.CreateExternalNode.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateExternalNode.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		plan.Name = types.String{Value: echoResp.CreateExternalNode.Name}
		if echoResp.CreateExternalNode.ReceiveMessageType != nil {
			plan.ReceiveMessageType = types.String{Value: echoResp.CreateExternalNode.ReceiveMessageType.Name}
		} else {
			plan.ReceiveMessageType = types.String{Null: true}
		}
		if echoResp.CreateExternalNode.SendMessageType != nil {
			plan.SendMessageType = types.String{Value: echoResp.CreateExternalNode.SendMessageType.Name}
		} else {
			plan.SendMessageType = types.String{Null: true}
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ExternalNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state externalNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ExternalNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ExternalNodeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := dataSendReceiveNodeSchema()
	for key, attribute := range schema {
		switch key {
		case "description":
			attribute.Computed = false
			attribute.Optional = true
		case "name":
			attribute.Computed = false
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
			attribute.Required = true
			attribute.Validators = common.NameValidators
		case "receive_message_type":
			attribute.Computed = false
			attribute.Optional = true
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
		case "send_message_type":
			attribute.Computed = false
			attribute.Optional = true
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
		}
		schema[key] = attribute
	}
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"app": {
				MarkdownDescription: "The ExternalApp or CrossAccountApp this Node is associated with.",
				Optional:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Type:                types.StringType,
			},
			"config": {
				MarkdownDescription: "The config, in JSON object format (i.e. - dict, map).",
				Optional:            true,
				Sensitive:           true,
				Type:                common.ConfigType{},
			},
		},
	)
	return tfsdk.Schema{
		Attributes: schema,
		MarkdownDescription: "[ExternalNodes](https://docs.echo.stream/docs/external-node) exist outside the " +
			"EchoStream Cloud. Can be part of an ExternalApp or CrossAccountApp. You may use any computing resource " +
			"or language that you want to implement them.",
	}, nil
}

func (r *ExternalNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ExternalNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_node"
}

func (r *ExternalNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state externalNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
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
				state.App = types.String{Value: app.Name}
			case *api.ExternalNodeFieldsAppExternalApp:
				state.App = types.String{Value: app.Name}
			default:
				resp.Diagnostics.AddError(
					"Invalid App type",
					fmt.Sprintf("Expected CrossAccountApp or ExternalApp, got %s", *app.GetTypename()),
				)
			}
			if node.Config != nil {
				state.Config = common.Config{Value: *node.Config}
			} else {
				state.Config = common.Config{Null: true}
			}
			if node.Description != nil {
				state.Description = types.String{Value: *node.Description}
			} else {
				state.Description = types.String{Null: true}
			}
			state.Name = types.String{Value: node.Name}
			if node.ReceiveMessageType != nil {
				state.ReceiveMessageType = types.String{Value: node.ReceiveMessageType.Name}
			} else {
				state.ReceiveMessageType = types.String{Null: true}
			}
			if node.SendMessageType != nil {
				state.SendMessageType = types.String{Value: node.SendMessageType.Name}
			} else {
				state.SendMessageType = types.String{Null: true}
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ExternalNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ExternalNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan externalNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config      *string
		description *string
	)
	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}

	if echoResp, err := api.UpdateExternalNode(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		config,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error updating ExternalNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find ExternalNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.Value))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateExternalNodeGetNodeExternalNode:
			switch app := (node.Update.App).(type) {
			case *api.ExternalNodeFieldsAppCrossAccountApp:
				plan.App = types.String{Value: app.Name}
			case *api.ExternalNodeFieldsAppExternalApp:
				plan.App = types.String{Value: app.Name}
			default:
				resp.Diagnostics.AddError(
					"Invalid App type",
					fmt.Sprintf("Expected CrossAccountApp or ExternalApp, got %s", *app.GetTypename()),
				)
			}
			if node.Update.Config != nil {
				plan.Config = common.Config{Value: *node.Update.Config}
			} else {
				plan.Config = common.Config{Null: true}
			}
			if node.Update.Description != nil {
				plan.Description = types.String{Value: *node.Update.Description}
			} else {
				plan.Description = types.String{Null: true}
			}
			plan.Name = types.String{Value: node.Update.Name}
			if node.Update.ReceiveMessageType != nil {
				plan.ReceiveMessageType = types.String{Value: node.Update.ReceiveMessageType.Name}
			} else {
				plan.ReceiveMessageType = types.String{Null: true}
			}
			if node.Update.SendMessageType != nil {
				plan.SendMessageType = types.String{Value: node.Update.SendMessageType.Name}
			} else {
				plan.SendMessageType = types.String{Null: true}
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ExternalNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
