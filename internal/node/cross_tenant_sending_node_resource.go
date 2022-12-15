package node

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigure        = &CrossTenantSendingNodeResource{}
	_ resource.ResourceWithConfigValidators = &CrossTenantSendingNodeResource{}
	_ resource.ResourceWithImportState      = &CrossTenantSendingNodeResource{}
	_ resource.ResourceWithModifyPlan       = &CrossTenantSendingNodeResource{}
)

// ProcessorNodeResource defines the resource implementation.
type CrossTenantSendingNodeResource struct {
	data *common.ProviderData
}

type crossTenantSendingNodeModel struct {
	App                  types.String  `tfsdk:"app"`
	Config               common.Config `tfsdk:"config"`
	Description          types.String  `tfsdk:"description"`
	InlineProcessor      types.String  `tfsdk:"inline_processor"`
	LoggingLevel         types.String  `tfsdk:"logging_level"`
	ManagedProcessor     types.String  `tfsdk:"managed_processor"`
	Name                 types.String  `tfsdk:"name"`
	ReceiveMessageType   types.String  `tfsdk:"receive_message_type"`
	Requirements         types.Set     `tfsdk:"requirements"`
	SendMessageType      types.String  `tfsdk:"send_message_type"`
	SequentialProcessing types.Bool    `tfsdk:"sequential_processing"`
}

func (r *CrossTenantSendingNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CrossTenantSendingNodeResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("inline_processor"),
			path.MatchRoot("managed_processor"),
		),
	}
}

func (r *CrossTenantSendingNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan crossTenantSendingNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config               *string
		description          *string
		diags                diag.Diagnostics
		inlineProcessor      *string
		loggingLevel         *api.LogLevel
		managedProcessor     *string
		requirements         []string
		sendMessageType      *string
		sequentialProcessing *bool
	)

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.InlineProcessor.IsNull() || plan.InlineProcessor.IsUnknown()) {
		temp := plan.InlineProcessor.ValueString()
		inlineProcessor = &temp
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		temp := plan.LoggingLevel.ValueString()
		loggingLevel = (*api.LogLevel)(&temp)
	}
	if !(plan.ManagedProcessor.IsNull() || plan.ManagedProcessor.IsUnknown()) {
		temp := plan.ManagedProcessor.ValueString()
		managedProcessor = &temp
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.SendMessageType.IsNull() || plan.SendMessageType.IsUnknown()) {
		temp := plan.SendMessageType.ValueString()
		sendMessageType = &temp
	}
	if !(plan.SequentialProcessing.IsNull() || plan.SequentialProcessing.IsUnknown()) {
		temp := plan.SequentialProcessing.ValueBool()
		sequentialProcessing = &temp
	}

	if echoResp, err := api.CreateCrossTenantSendingNode(
		ctx,
		r.data.Client,
		plan.App.ValueString(),
		plan.Name.ValueString(),
		plan.ReceiveMessageType.ValueString(),
		r.data.Tenant,
		config,
		description,
		inlineProcessor,
		loggingLevel,
		managedProcessor,
		requirements,
		sendMessageType,
		sequentialProcessing,
	); err != nil {
		resp.Diagnostics.AddError("Error creating CrossTenantSendingNode", err.Error())
		return
	} else {
		plan.App = types.StringValue(echoResp.CreateCrossTenantSendingNode.App.Name)
		if echoResp.CreateCrossTenantSendingNode.Config != nil {
			plan.Config = common.ConfigValue(*echoResp.CreateCrossTenantSendingNode.Config)
		} else {
			plan.Config = common.ConfigNull()
		}
		if echoResp.CreateCrossTenantSendingNode.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateCrossTenantSendingNode.Description)
		} else {
			plan.Description = types.StringNull()
		}
		if echoResp.CreateCrossTenantSendingNode.InlineProcessor != nil {
			plan.InlineProcessor = types.StringValue(*echoResp.CreateCrossTenantSendingNode.InlineProcessor)
		} else {
			plan.InlineProcessor = types.StringNull()
		}
		if echoResp.CreateCrossTenantSendingNode.LoggingLevel != nil {
			plan.LoggingLevel = types.StringValue(string(*echoResp.CreateCrossTenantSendingNode.LoggingLevel))
		} else {
			plan.LoggingLevel = types.StringNull()
		}
		if echoResp.CreateCrossTenantSendingNode.ManagedProcessor != nil {
			plan.ManagedProcessor = types.StringValue(echoResp.CreateCrossTenantSendingNode.ManagedProcessor.Name)
		} else {
			plan.ManagedProcessor = types.StringNull()
		}
		plan.Name = types.StringValue(echoResp.CreateCrossTenantSendingNode.Name)
		plan.ReceiveMessageType = types.StringValue(echoResp.CreateCrossTenantSendingNode.ReceiveMessageType.Name)
		if len(echoResp.CreateCrossTenantSendingNode.Requirements) > 0 {
			elems := []attr.Value{}
			for _, req := range echoResp.CreateCrossTenantSendingNode.Requirements {
				elems = append(elems, types.StringValue(req))
			}
			plan.Requirements, diags = types.SetValue(types.StringType, elems)
			if diags != nil && diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
		} else {
			plan.Requirements = types.SetNull(types.StringType)
		}
		if echoResp.CreateCrossTenantSendingNode.SendMessageType != nil {
			plan.SendMessageType = types.StringValue(echoResp.CreateCrossTenantSendingNode.SendMessageType.Name)
		} else {
			plan.SendMessageType = types.StringNull()
		}
		if echoResp.CreateCrossTenantSendingNode.SequentialProcessing != nil {
			plan.SequentialProcessing = types.BoolValue(*echoResp.CreateCrossTenantSendingNode.SequentialProcessing)
		} else {
			plan.SequentialProcessing = types.BoolValue(false)
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CrossTenantSendingNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state crossTenantSendingNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting CrossTenantSendingNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *CrossTenantSendingNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *CrossTenantSendingNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cross_tenant_sending_node"
}

func (r *CrossTenantSendingNodeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state crossTenantSendingNodeModel

	// If the entire state is null, resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// If the entire plan is null, the resource is planned for destruction.
	if req.Plan.Raw.IsNull() {
		resp.Diagnostics.AddWarning(
			"Deleting CrossTenantSendingNode",
			"This will also delete the receiving Tenant's CrossTenantReceivingNode!",
		)
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var plan crossTenantSendingNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if !plan.App.Equal(state.App) {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("name"),
			"Changing a CrossTenantSendingNode's `name`",
			"This will result in the deletion of the receiving Tenant's CrossTenantReceivingNode.",
		)
	}
	if !plan.Name.Equal(state.Name) {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("name"),
			"Changing a CrossTenantSendingNode's `name`",
			"This will result in the deletion of the receiving Tenant's CrossTenantReceivingNode.",
		)
	}
	if !plan.ReceiveMessageType.Equal(state.ReceiveMessageType) {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("receive_message_type"),
			"Changing a CrossTenantSendingNode's `receive_message_type`",
			"This will result in the deletion of the receiving Tenant's CrossTenantReceivingNode.",
		)
	}
	if !plan.SendMessageType.Equal(state.SendMessageType) {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("send_message_type"),
			"Changing a CrossTenantSendingNode's `send_message_type`",
			"This will result in the deletion of the receiving Tenant's CrossTenantReceivingNode.",
		)
	}
}

func (r *CrossTenantSendingNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		state crossTenantSendingNodeModel
	)

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossTenantSendingNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeCrossTenantSendingNode:
			state.App = types.StringValue(node.App.Name)
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
			if node.InlineProcessor != nil {
				state.InlineProcessor = types.StringValue(*node.InlineProcessor)
			} else {
				state.InlineProcessor = types.StringNull()
			}
			if node.LoggingLevel != nil {
				state.LoggingLevel = types.StringValue(string(*node.LoggingLevel))
			} else {
				state.LoggingLevel = types.StringNull()
			}
			if node.ManagedProcessor != nil {
				state.ManagedProcessor = types.StringValue(node.ManagedProcessor.Name)
			} else {
				state.ManagedProcessor = types.StringNull()
			}
			state.Name = types.StringValue(node.Name)
			state.ReceiveMessageType = types.StringValue(node.ReceiveMessageType.Name)
			if len(node.Requirements) > 0 {
				elems := []attr.Value{}
				for _, req := range node.Requirements {
					elems = append(elems, types.StringValue(req))
				}
				state.Requirements, diags = types.SetValue(types.StringType, elems)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			} else {
				state.Requirements = types.SetNull(types.StringType)
			}
			if node.SendMessageType != nil {
				state.SendMessageType = types.StringValue(node.SendMessageType.Name)
			} else {
				state.SendMessageType = types.StringNull()
			}
			if node.SequentialProcessing != nil {
				state.SequentialProcessing = types.BoolValue(*node.SequentialProcessing)
			} else {
				state.SequentialProcessing = types.BoolValue(false)
			}
		default:
			resp.Diagnostics.AddError(
				"Expected CrossTenantSendingNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CrossTenantSendingNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				MarkdownDescription: "The CrossTenantSendingApp this Node is associated with.",
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
			"inline_processor": schema.StringAttribute{
				MarkdownDescription: "A Python code string that contains a single top-level function definition." +
					"This function is used as a template when creating custom processing in ProcessorNodes" +
					"that use this MessageType. This function must have the signature" +
					"`(*, context, message, source, **kwargs)` and return None, a string or a list of strings." +
					" Mutually exclusive with `managedProcessor`.",
				Optional: true,
			},
			"logging_level": schema.StringAttribute{
				MarkdownDescription: "The logging level. One of `DEBUG`, `ERROR`, `INFO`, `WARNING`. Defaults to `INFO`.",
				Optional:            true,
				Validators:          []validator.String{common.LogLevelValidator},
			},
			"managed_processor": schema.StringAttribute{
				MarkdownDescription: "The managedProcessor. Mutually exclusive with the `inlineProcessor`.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Node. Must be unique within the Tenant.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
				Validators:          common.FunctionNodeNameValidators,
			},
			"receive_message_type": schema.StringAttribute{
				MarkdownDescription: "The MessageType that this Node is capable of receiving.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
			},
			"requirements": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format.",
				Optional:            true,
				Validators:          []validator.Set{common.RequirementsValidator},
			},
			"send_message_type": schema.StringAttribute{
				MarkdownDescription: "The MessageType that this Node is capable of sending.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Optional:            true,
			},
			"sequential_processing": schema.BoolAttribute{
				MarkdownDescription: "`true` if messages should not be processed concurrently. If `false`, messages are processed concurrently. Defaults to `false`.",
				Optional:            true,
			},
		},
		MarkdownDescription: "[CrossTenantSendingNodes](https://docs.echo.stream/docs/cross-tenant-sending-node) send messages to a receiving Tenant.",
	}
}

func (r *CrossTenantSendingNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan crossTenantSendingNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config               *string
		description          *string
		diags                diag.Diagnostics
		inlineProcessor      *string
		loggingLevel         *api.LogLevel
		managedProcessor     *string
		requirements         []string
		sequentialProcessing *bool
	)
	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.InlineProcessor.IsNull() || plan.InlineProcessor.IsUnknown()) {
		temp := plan.InlineProcessor.ValueString()
		inlineProcessor = &temp
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		temp := plan.LoggingLevel.ValueString()
		loggingLevel = (*api.LogLevel)(&temp)
	}
	if !(plan.ManagedProcessor.IsNull() || plan.ManagedProcessor.IsUnknown()) {
		temp := plan.ManagedProcessor.ValueString()
		managedProcessor = &temp
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.SequentialProcessing.IsNull() || plan.SequentialProcessing.IsUnknown()) {
		temp := plan.SequentialProcessing.ValueBool()
		sequentialProcessing = &temp
	}

	if echoResp, err := api.UpdateCrossTenantSendingNode(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		r.data.Tenant,
		config,
		description,
		inlineProcessor,
		loggingLevel,
		managedProcessor,
		requirements,
		sequentialProcessing,
	); err != nil {
		resp.Diagnostics.AddError("Error updating CrossTenantSendingNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find CrossTenantSendingNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.ValueString()))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateCrossTenantSendingNodeGetNodeCrossTenantSendingNode:
			plan.App = types.StringValue(node.Update.App.Name)
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
			if node.Update.InlineProcessor != nil {
				plan.InlineProcessor = types.StringValue(*node.Update.InlineProcessor)
			} else {
				plan.InlineProcessor = types.StringNull()
			}
			if node.Update.LoggingLevel != nil {
				plan.LoggingLevel = types.StringValue(string(*node.Update.LoggingLevel))
			} else {
				plan.LoggingLevel = types.StringNull()
			}
			if node.Update.ManagedProcessor != nil {
				plan.ManagedProcessor = types.StringValue(node.Update.ManagedProcessor.Name)
			} else {
				plan.ManagedProcessor = types.StringNull()
			}
			plan.Name = types.StringValue(node.Update.Name)
			plan.ReceiveMessageType = types.StringValue(node.Update.ReceiveMessageType.Name)
			if len(node.Update.Requirements) > 0 {
				elems := []attr.Value{}
				for _, req := range node.Update.Requirements {
					elems = append(elems, types.StringValue(req))
				}
				plan.Requirements, diags = types.SetValue(types.StringType, elems)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			} else {
				plan.Requirements = types.SetNull(types.StringType)
			}
			if node.Update.SendMessageType != nil {
				plan.SendMessageType = types.StringValue(node.Update.SendMessageType.Name)
			} else {
				plan.SendMessageType = types.StringNull()
			}
			if node.Update.SequentialProcessing != nil {
				plan.SequentialProcessing = types.BoolValue(*node.Update.SequentialProcessing)
			} else {
				plan.SequentialProcessing = types.BoolValue(false)
			}
		default:
			resp.Diagnostics.AddError(
				"Expected CrossTenantSendingNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
