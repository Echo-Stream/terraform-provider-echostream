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
	_ resource.ResourceWithConfigValidators = &ProcessorNodeResource{}
	_ resource.ResourceWithImportState      = &ProcessorNodeResource{}
	_ resource.ResourceWithSchema           = &ProcessorNodeResource{}
)

// ProcessorNodeResource defines the resource implementation.
type ProcessorNodeResource struct {
	data *common.ProviderData
}

type processorNodeModel struct {
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

func (r *ProcessorNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProcessorNodeResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("inline_processor"),
			path.MatchRoot("managed_processor"),
		),
	}
}

func (r *ProcessorNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan processorNodeModel

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

	if echoResp, err := api.CreateProcessorNode(
		ctx,
		r.data.Client,
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
		resp.Diagnostics.AddError("Error creating ProcessorNode", err.Error())
		return
	} else {
		if echoResp.CreateProcessorNode.Config != nil {
			plan.Config = common.ConfigValue(*echoResp.CreateProcessorNode.Config)
		} else {
			plan.Config = common.ConfigNull()
		}
		if echoResp.CreateProcessorNode.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateProcessorNode.Description)
		} else {
			plan.Description = types.StringNull()
		}
		if echoResp.CreateProcessorNode.InlineProcessor != nil {
			plan.InlineProcessor = types.StringValue(*echoResp.CreateProcessorNode.InlineProcessor)
		} else {
			plan.InlineProcessor = types.StringNull()
		}
		if echoResp.CreateProcessorNode.LoggingLevel != nil {
			plan.LoggingLevel = types.StringValue(string(*echoResp.CreateProcessorNode.LoggingLevel))
		} else {
			plan.LoggingLevel = types.StringNull()
		}
		if echoResp.CreateProcessorNode.ManagedProcessor != nil {
			plan.ManagedProcessor = types.StringValue(echoResp.CreateProcessorNode.ManagedProcessor.Name)
		} else {
			plan.ManagedProcessor = types.StringNull()
		}
		plan.Name = types.StringValue(echoResp.CreateProcessorNode.Name)
		plan.ReceiveMessageType = types.StringValue(echoResp.CreateProcessorNode.ReceiveMessageType.Name)
		if len(echoResp.CreateProcessorNode.Requirements) > 0 {
			elems := []attr.Value{}
			for _, req := range echoResp.CreateProcessorNode.Requirements {
				elems = append(elems, types.StringValue(req))
			}
			plan.Requirements, diags = types.SetValue(types.StringType, elems)
			if diags != nil && diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
		} else {
			plan.Requirements = types.SetNull(types.StringType)
		}
		if echoResp.CreateProcessorNode.SendMessageType != nil {
			plan.SendMessageType = types.StringValue(echoResp.CreateProcessorNode.SendMessageType.Name)
		} else {
			plan.SendMessageType = types.StringNull()
		}
		if echoResp.CreateProcessorNode.SequentialProcessing != nil {
			plan.SequentialProcessing = types.BoolValue(*echoResp.CreateProcessorNode.SequentialProcessing)
		} else {
			plan.SequentialProcessing = types.BoolValue(false)
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProcessorNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state processorNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ProcessorNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ProcessorNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ProcessorNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_processor_node"
}

func (r *ProcessorNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		state processorNodeModel
	)

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ProcessorNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeProcessorNode:
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
				"Expected ProcessorNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProcessorNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"sequential_processing": schema.BoolAttribute{
				MarkdownDescription: "`true` if messages should not be processed concurrently. If `false`, messages are processed concurrently. Defaults to `false`.",
				Optional:            true,
			},
		},
		MarkdownDescription: "[ProcessorNodes](https://docs.echo.stream/docs/processor-node) allow for almost any processing of messages, " +
			"including transformation, augmentation, generation, combination and splitting.",
	}
}

func (r *ProcessorNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan processorNodeModel

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

	if echoResp, err := api.UpdateProcessorNode(
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
		resp.Diagnostics.AddError("Error updating ProcessorNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find ProcessorNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.ValueString()))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateProcessorNodeGetNodeProcessorNode:
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
				"Expected ProcessorNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
