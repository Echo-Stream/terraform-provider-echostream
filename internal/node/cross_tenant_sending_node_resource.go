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

	if echoResp, err := api.CreateCrossTenantSendingNode(
		ctx,
		r.data.Client,
		plan.App.Value,
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
		resp.Diagnostics.AddError("Error creating CrossTenantSendingNode", err.Error())
		return
	} else {
		plan.App = types.String{Value: echoResp.CreateCrossTenantSendingNode.App.Name}
		if echoResp.CreateCrossTenantSendingNode.Config != nil {
			plan.Config = common.Config{Value: *echoResp.CreateCrossTenantSendingNode.Config}
		} else {
			plan.Config = common.Config{Null: true}
		}
		if echoResp.CreateCrossTenantSendingNode.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateCrossTenantSendingNode.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		if echoResp.CreateCrossTenantSendingNode.InlineProcessor != nil {
			plan.InlineProcessor = types.String{Value: *echoResp.CreateCrossTenantSendingNode.InlineProcessor}
		} else {
			plan.InlineProcessor = types.String{Null: true}
		}
		if echoResp.CreateCrossTenantSendingNode.LoggingLevel != nil {
			plan.LoggingLevel = types.String{Value: string(*echoResp.CreateCrossTenantSendingNode.LoggingLevel)}
		} else {
			plan.LoggingLevel = types.String{Null: true}
		}
		if echoResp.CreateCrossTenantSendingNode.ManagedProcessor != nil {
			plan.ManagedProcessor = types.String{Value: echoResp.CreateCrossTenantSendingNode.ManagedProcessor.Name}
		} else {
			plan.ManagedProcessor = types.String{Null: true}
		}
		plan.Name = types.String{Value: echoResp.CreateCrossTenantSendingNode.Name}
		plan.ReceiveMessageType = types.String{Value: echoResp.CreateCrossTenantSendingNode.ReceiveMessageType.Name}
		if len(echoResp.CreateCrossTenantSendingNode.Requirements) > 0 {
			plan.Requirements = types.Set{ElemType: types.StringType}
			for _, req := range echoResp.CreateCrossTenantSendingNode.Requirements {
				plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
			}
		} else {
			plan.Requirements = types.Set{ElemType: types.StringType, Null: true}
		}
		if echoResp.CreateCrossTenantSendingNode.SendMessageType != nil {
			plan.SendMessageType = types.String{Value: echoResp.CreateCrossTenantSendingNode.SendMessageType.Name}
		} else {
			plan.SendMessageType = types.String{Null: true}
		}
		if echoResp.CreateCrossTenantSendingNode.SequentialProcessing != nil {
			plan.SequentialProcessing = types.Bool{Value: *echoResp.CreateCrossTenantSendingNode.SequentialProcessing}
		} else {
			plan.SequentialProcessing = types.Bool{Value: false}
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

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting CrossTenantSendingNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *CrossTenantSendingNodeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"app": {
				Description:         "",
				MarkdownDescription: "",
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Required:            true,
				Type:                types.StringType,
			},
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
		Description:         "CrossTenantSendingNodes send messages to a receiving Tenant",
		MarkdownDescription: "CrossTenantSendingNodes send messages to a receiving Tenant",
	}, nil
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
	var state crossTenantSendingNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossTenantSendingNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeCrossTenantSendingNode:
			state.App = types.String{Value: node.App.Name}
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
			if len(node.Requirements) > 0 {
				state.Requirements = types.Set{ElemType: types.StringType}
				for _, req := range node.Requirements {
					state.Requirements.Elems = append(state.Requirements.Elems, types.String{Value: req})
				}
			} else {
				state.Requirements = types.Set{ElemType: types.StringType, Null: true}
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
				"Expected CrossTenantSendingNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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

	if echoResp, err := api.UpdateCrossTenantSendingNode(
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
		resp.Diagnostics.AddError("Error updating CrossTenantSendingNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find CrossTenantSendingNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.Value))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateCrossTenantSendingNodeGetNodeCrossTenantSendingNode:
			plan.App = types.String{Value: node.Update.App.Name}
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
			if len(node.Update.Requirements) > 0 {
				plan.Requirements = types.Set{ElemType: types.StringType}
				for _, req := range node.Update.Requirements {
					plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
				}
			} else {
				plan.Requirements = types.Set{ElemType: types.StringType, Null: true}
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
				"Expected CrossTenantSendingNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
