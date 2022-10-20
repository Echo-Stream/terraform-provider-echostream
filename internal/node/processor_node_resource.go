package node

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigValidators = &ProcessorNodeResource{}
	_ resource.ResourceWithImportState      = &ProcessorNodeResource{}
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
		inlineProcessor      *string
		loggingLevel         *api.LogLevel
		managedProcessor     *string
		requirements         []string
		sendMessageType      *string
		sequentialProcessing *bool
	)

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.InlineProcessor.IsNull() || plan.InlineProcessor.IsUnknown()) {
		inlineProcessor = &plan.InlineProcessor.Value
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		loggingLevel = (*api.LogLevel)(&plan.LoggingLevel.Value)
	}
	if !(plan.ManagedProcessor.IsNull() || plan.ManagedProcessor.IsUnknown()) {
		managedProcessor = &plan.ManagedProcessor.Value
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.SendMessageType.IsNull() || plan.SendMessageType.IsUnknown()) {
		sendMessageType = &plan.SendMessageType.Value
	}
	if !(plan.SequentialProcessing.IsNull() || plan.SequentialProcessing.IsUnknown()) {
		sequentialProcessing = &plan.SequentialProcessing.Value
	}

	if echoResp, err := api.CreateProcessorNode(
		ctx,
		r.data.Client,
		plan.Name.Value,
		plan.ReceiveMessageType.Value,
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
			plan.Config = common.Config{Value: *echoResp.CreateProcessorNode.Config}
		} else {
			plan.Config = common.Config{Null: true}
		}
		if echoResp.CreateProcessorNode.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateProcessorNode.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		if echoResp.CreateProcessorNode.InlineProcessor != nil {
			plan.InlineProcessor = types.String{Value: *echoResp.CreateProcessorNode.InlineProcessor}
		} else {
			plan.InlineProcessor = types.String{Null: true}
		}
		if echoResp.CreateProcessorNode.LoggingLevel != nil {
			plan.LoggingLevel = types.String{Value: string(*echoResp.CreateProcessorNode.LoggingLevel)}
		} else {
			plan.LoggingLevel = types.String{Null: true}
		}
		if echoResp.CreateProcessorNode.ManagedProcessor != nil {
			plan.ManagedProcessor = types.String{Value: echoResp.CreateProcessorNode.ManagedProcessor.Name}
		} else {
			plan.ManagedProcessor = types.String{Null: true}
		}
		plan.Name = types.String{Value: echoResp.CreateProcessorNode.Name}
		plan.ReceiveMessageType = types.String{Value: echoResp.CreateProcessorNode.ReceiveMessageType.Name}
		plan.Requirements = types.Set{ElemType: types.StringType}
		if len(echoResp.CreateProcessorNode.Requirements) > 0 {
			for _, req := range echoResp.CreateProcessorNode.Requirements {
				plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
			}
		} else {
			plan.Requirements.Null = true
		}
		if echoResp.CreateProcessorNode.SendMessageType != nil {
			plan.SendMessageType = types.String{Value: echoResp.CreateProcessorNode.SendMessageType.Name}
		} else {
			plan.SendMessageType = types.String{Null: true}
		}
		if echoResp.CreateProcessorNode.SequentialProcessing != nil {
			plan.SequentialProcessing = types.Bool{Value: *echoResp.CreateProcessorNode.SequentialProcessing}
		} else {
			plan.SequentialProcessing = types.Bool{Value: false}
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

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ProcessorNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ProcessorNodeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			attribute.Validators = common.FunctionNodeNameValidators
		case "receive_message_type":
			attribute.Computed = false
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
			attribute.Required = true
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
			"config": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Sensitive:           true,
				Type:                common.ConfigType{},
			},
			"inline_processor": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
			},
			"logging_level": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
				Validators:          []tfsdk.AttributeValidator{common.LogLevelValidator},
			},
			"managed_processor": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
			},
			"requirements": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.SetType{ElemType: types.StringType},
				Validators:          []tfsdk.AttributeValidator{common.RequirementsValidator},
			},
			"sequential_processing": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.BoolType,
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		Description:         "ProcessorNodes allow for processing messages",
		MarkdownDescription: "ProcessorNodes allow for processing messages",
	}, nil
}

func (r *ProcessorNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ProcessorNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_processor_node"
}

func (r *ProcessorNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state processorNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ProcessorNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeProcessorNode:
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
			if node.InlineProcessor != nil {
				state.InlineProcessor = types.String{Value: *node.InlineProcessor}
			} else {
				state.InlineProcessor = types.String{Null: true}
			}
			if node.LoggingLevel != nil {
				state.LoggingLevel = types.String{Value: string(*node.LoggingLevel)}
			} else {
				state.LoggingLevel = types.String{Null: true}
			}
			if node.ManagedProcessor != nil {
				state.ManagedProcessor = types.String{Value: node.ManagedProcessor.Name}
			} else {
				state.ManagedProcessor = types.String{Null: true}
			}
			state.Name = types.String{Value: node.Name}
			state.ReceiveMessageType = types.String{Value: node.ReceiveMessageType.Name}
			state.Requirements = types.Set{ElemType: types.StringType}
			if len(node.Requirements) > 0 {
				for _, req := range node.Requirements {
					state.Requirements.Elems = append(state.Requirements.Elems, types.String{Value: req})
				}
			} else {
				state.Requirements.Null = true
			}
			if node.SendMessageType != nil {
				state.SendMessageType = types.String{Value: node.SendMessageType.Name}
			} else {
				state.SendMessageType = types.String{Null: true}
			}
			if node.SequentialProcessing != nil {
				state.SequentialProcessing = types.Bool{Value: *node.SequentialProcessing}
			} else {
				state.SequentialProcessing = types.Bool{Value: false}
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ProcessorNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
		inlineProcessor      *string
		loggingLevel         *api.LogLevel
		managedProcessor     *string
		requirements         []string
		sequentialProcessing *bool
	)
	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.InlineProcessor.IsNull() || plan.InlineProcessor.IsUnknown()) {
		inlineProcessor = &plan.InlineProcessor.Value
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		loggingLevel = (*api.LogLevel)(&plan.LoggingLevel.Value)
	}
	if !(plan.ManagedProcessor.IsNull() || plan.ManagedProcessor.IsUnknown()) {
		managedProcessor = &plan.ManagedProcessor.Value
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.SequentialProcessing.IsNull() || plan.SequentialProcessing.IsUnknown()) {
		sequentialProcessing = &plan.SequentialProcessing.Value
	}

	if echoResp, err := api.UpdateProcessorNode(
		ctx,
		r.data.Client,
		plan.Name.Value,
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
		resp.Diagnostics.AddError("Cannot find ProcessorNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.Value))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateProcessorNodeGetNodeProcessorNode:
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
			if node.Update.InlineProcessor != nil {
				plan.InlineProcessor = types.String{Value: *node.Update.InlineProcessor}
			} else {
				plan.InlineProcessor = types.String{Null: true}
			}
			if node.Update.LoggingLevel != nil {
				plan.LoggingLevel = types.String{Value: string(*node.Update.LoggingLevel)}
			} else {
				plan.LoggingLevel = types.String{Null: true}
			}
			if node.Update.ManagedProcessor != nil {
				plan.ManagedProcessor = types.String{Value: node.Update.ManagedProcessor.Name}
			} else {
				plan.ManagedProcessor = types.String{Null: true}
			}
			plan.Name = types.String{Value: node.Update.Name}
			plan.ReceiveMessageType = types.String{Value: node.Update.ReceiveMessageType.Name}
			plan.Requirements = types.Set{ElemType: types.StringType}
			if len(node.Update.Requirements) > 0 {
				for _, req := range node.Update.Requirements {
					plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
				}
			} else {
				plan.Requirements.Null = true
			}
			if node.Update.SendMessageType != nil {
				plan.SendMessageType = types.String{Value: node.Update.SendMessageType.Name}
			} else {
				plan.SendMessageType = types.String{Null: true}
			}
			if node.Update.SequentialProcessing != nil {
				plan.SequentialProcessing = types.Bool{Value: *node.Update.SequentialProcessing}
			} else {
				plan.SequentialProcessing = types.Bool{Value: false}
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ProcessorNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
