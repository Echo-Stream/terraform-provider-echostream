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
	_ resource.ResourceWithConfigure   = &ExternalNodeResource{}
	_ resource.ResourceWithImportState = &ExternalNodeResource{}
)

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
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.ReceiveMessageType.IsNull() || plan.ReceiveMessageType.IsUnknown()) {
		temp := plan.ReceiveMessageType.ValueString()
		receiveMessageType = &temp
	}
	if !(plan.SendMessageType.IsNull() || plan.SendMessageType.IsUnknown()) {
		temp := plan.SendMessageType.ValueString()
		sendMessageType = &temp
	}

	if echoResp, err := api.CreateExternalNode(
		ctx,
		r.data.Client,
		plan.App.ValueString(),
		plan.Name.ValueString(),
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
			plan.App = types.StringValue(app.Name)
		case *api.ExternalNodeFieldsAppExternalApp:
			plan.App = types.StringValue(app.Name)
		default:
			resp.Diagnostics.AddError(
				"Invalid App type",
				fmt.Sprintf("Expected CrossAccountApp or ExternalApp, got %s", *app.GetTypename()),
			)
		}
		if echoResp.CreateExternalNode.Config != nil {
			plan.Config = common.ConfigValue(*echoResp.CreateExternalNode.Config)
		} else {
			plan.Config = common.ConfigNull()
		}
		if echoResp.CreateExternalNode.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateExternalNode.Description)
		} else {
			plan.Description = types.StringNull()
		}
		plan.Name = types.StringValue(echoResp.CreateExternalNode.Name)
		if echoResp.CreateExternalNode.ReceiveMessageType != nil {
			plan.ReceiveMessageType = types.StringValue(echoResp.CreateExternalNode.ReceiveMessageType.Name)
		} else {
			plan.ReceiveMessageType = types.StringNull()
		}
		if echoResp.CreateExternalNode.SendMessageType != nil {
			plan.SendMessageType = types.StringValue(echoResp.CreateExternalNode.SendMessageType.Name)
		} else {
			plan.SendMessageType = types.StringNull()
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

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ExternalNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
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

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
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
				state.App = types.StringValue(app.Name)
			case *api.ExternalNodeFieldsAppExternalApp:
				state.App = types.StringValue(app.Name)
			default:
				resp.Diagnostics.AddError(
					"Invalid App type",
					fmt.Sprintf("Expected CrossAccountApp or ExternalApp, got %s", *app.GetTypename()),
				)
			}
			if node.Config != nil {
				state.Config = common.ConfigValue(*node.Config)
			} else {
				state.Config = common.ConfigNull()
			}
			if node.Description != nil {
				state.Description = types.StringValue(*node.Description)
			} else {
				state.Description = types.StringNull()
			}
			state.Name = types.StringValue(node.Name)
			if node.ReceiveMessageType != nil {
				state.ReceiveMessageType = types.StringValue(node.ReceiveMessageType.Name)
			} else {
				state.ReceiveMessageType = types.StringNull()
			}
			if node.SendMessageType != nil {
				state.SendMessageType = types.StringValue(node.SendMessageType.Name)
			} else {
				state.SendMessageType = types.StringNull()
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ExternalNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ExternalNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				MarkdownDescription: "The ExternalApp or CrossAccountApp this Node is associated with.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
			},
			"config": schema.StringAttribute{
				CustomType:          common.ConfigType{},
				MarkdownDescription: "The config, in JSON object format (i.e. - dict, map).",
				Optional:            true,
				Sensitive:           true,
			},
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
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"send_message_type": schema.StringAttribute{
				MarkdownDescription: "The MessageType that this Node is capable of sending.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
		MarkdownDescription: "[ExternalNodes](https://docs.echo.stream/docs/external-node) exist outside the " +
			"EchoStream Cloud. Can be part of an ExternalApp or CrossAccountApp. You may use any computing resource " +
			"or language that you want to implement them.",
	}
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
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}

	if echoResp, err := api.UpdateExternalNode(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		r.data.Tenant,
		config,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error updating ExternalNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find ExternalNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.ValueString()))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateExternalNodeGetNodeExternalNode:
			switch app := (node.Update.App).(type) {
			case *api.ExternalNodeFieldsAppCrossAccountApp:
				plan.App = types.StringValue(app.Name)
			case *api.ExternalNodeFieldsAppExternalApp:
				plan.App = types.StringValue(app.Name)
			default:
				resp.Diagnostics.AddError(
					"Invalid App type",
					fmt.Sprintf("Expected CrossAccountApp or ExternalApp, got %s", *app.GetTypename()),
				)
			}
			if node.Update.Config != nil {
				plan.Config = common.ConfigValue(*node.Update.Config)
			} else {
				plan.Config = common.ConfigNull()
			}
			if node.Update.Description != nil {
				plan.Description = types.StringValue(*node.Update.Description)
			} else {
				plan.Description = types.StringNull()
			}
			plan.Name = types.StringValue(node.Update.Name)
			if node.Update.ReceiveMessageType != nil {
				plan.ReceiveMessageType = types.StringValue(node.Update.ReceiveMessageType.Name)
			} else {
				plan.ReceiveMessageType = types.StringNull()
			}
			if node.Update.SendMessageType != nil {
				plan.SendMessageType = types.StringValue(node.Update.SendMessageType.Name)
			} else {
				plan.SendMessageType = types.StringNull()
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ExternalNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
