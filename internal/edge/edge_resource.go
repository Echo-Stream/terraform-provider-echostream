package edge

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigure   = &EdgeResource{}
	_ resource.ResourceWithImportState = &EdgeResource{}
	_ resource.ResourceWithModifyPlan  = &EdgeResource{}
)

// EdgeResource defines the resource implementation.
type EdgeResource struct {
	data *common.ProviderData
}

func (r *EdgeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type edgeModel struct {
	Arn             types.String `tfsdk:"arn"`
	Description     types.String `tfsdk:"description"`
	KmsKey          types.String `tfsdk:"kmskey"`
	MaxReceiveCount types.Int64  `tfsdk:"max_receive_count"`
	MessageType     types.String `tfsdk:"message_type"`
	Queue           types.String `tfsdk:"queue"`
	Source          types.String `tfsdk:"source"`
	Target          types.String `tfsdk:"target"`
}

func (r *EdgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan edgeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		description     *string
		kmsKey          *string
		maxReceiveCount *int
	)
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.KmsKey.IsNull() || plan.KmsKey.IsUnknown()) {
		temp := plan.KmsKey.ValueString()
		kmsKey = &temp
	}
	if !(plan.MaxReceiveCount.IsNull() || plan.MaxReceiveCount.IsUnknown()) {
		temp := int(plan.MaxReceiveCount.ValueInt64())
		maxReceiveCount = &temp
	}

	if echoResp, err := api.CreateEdge(
		ctx,
		r.data.Client,
		plan.Source.ValueString(),
		plan.Target.ValueString(),
		r.data.Tenant,
		description,
		kmsKey,
		maxReceiveCount,
	); err != nil {
		resp.Diagnostics.AddError("Error creating Edge", err.Error())
		return
	} else {
		plan.Arn = types.StringValue(echoResp.CreateEdge.Arn)
		if echoResp.CreateEdge.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateEdge.Description)
		} else {
			plan.Description = types.StringNull()
		}
		if echoResp.CreateEdge.KmsKey != nil {
			plan.KmsKey = types.StringValue(echoResp.CreateEdge.KmsKey.Name)
		} else {
			plan.KmsKey = types.StringNull()
		}
		if echoResp.CreateEdge.MaxReceiveCount != nil {
			plan.MaxReceiveCount = types.Int64Value(int64(*echoResp.CreateEdge.MaxReceiveCount))
		} else {
			plan.MaxReceiveCount = types.Int64Null()
		}
		plan.MessageType = types.StringValue(echoResp.CreateEdge.MessageType.Name)
		plan.Queue = types.StringValue(echoResp.CreateEdge.Queue)
		plan.Source = types.StringValue(echoResp.CreateEdge.Source.GetName())
		plan.Target = types.StringValue(echoResp.CreateEdge.Target.GetName())
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EdgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state edgeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteEdge(ctx, r.data.Client, state.Source.ValueString(), state.Target.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting Edge", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *EdgeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	nodes := strings.Split(req.ID, "|")
	if len(nodes) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Edge ID",
			fmt.Sprintf("Edge ID must be <source>|<target>, found %s", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source"), nodes[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target"), nodes[1])...)
}

func (r *EdgeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge"
}

func (r *EdgeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var (
		plan  edgeModel
		state edgeModel
	)

	// If the entire state is null or the entire plan is null, resource is being created or destroyed.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Read Terraform state and plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	newSourceMessageType := state.MessageType.ValueString()
	newTargetMessageType := state.MessageType.ValueString()

	if !state.Source.Equal(plan.Source) {
		if echoResp, err := api.ReadNodeMessageTypes(ctx, r.data.Client, plan.Source.ValueString(), r.data.Tenant); err != nil {
			resp.Diagnostics.AddError("Error reading planned source", err.Error())
			return
		} else if echoResp.GetNode == nil {
			resp.Diagnostics.AddAttributeError(path.Root("source"), "Cannot find Node", fmt.Sprintf("'%s' Node does not exist", plan.Source.ValueString()))
			return
		} else {
			node := reflect.Indirect(reflect.ValueOf(*echoResp.GetNode))
			if smt := reflect.Indirect(node.FieldByName("SendMessageType")); !smt.IsZero() {
				if name := smt.FieldByName("Name").String(); name != state.MessageType.ValueString() {
					newSourceMessageType = name
					resp.RequiresReplace = append(resp.RequiresReplace, path.Root("source"))
				}
			} else {
				resp.Diagnostics.AddAttributeError(path.Root("source"), "Invalid planned source", fmt.Sprintf("'%s' Node does not send messages", plan.Source.ValueString()))
			}
		}
	}
	if !state.Target.Equal(plan.Target) {
		if echoResp, err := api.ReadNodeMessageTypes(ctx, r.data.Client, plan.Target.ValueString(), r.data.Tenant); err != nil {
			resp.Diagnostics.AddError("Error reading planned target", err.Error())
			return
		} else if echoResp.GetNode == nil {
			resp.Diagnostics.AddAttributeError(path.Root("target"), "Cannot find Node", fmt.Sprintf("'%s' Node does not exist", plan.Target.ValueString()))
			return
		} else {
			node := reflect.Indirect(reflect.ValueOf(*echoResp.GetNode))
			if smt := reflect.Indirect(node.FieldByName("ReceiveMessageType")); !smt.IsZero() {
				if name := smt.FieldByName("Name").String(); name != state.MessageType.ValueString() {
					newTargetMessageType = name
					resp.RequiresReplace = append(resp.RequiresReplace, path.Root("target"))
				}
			} else {
				resp.Diagnostics.AddAttributeError(path.Root("source"), "Invalid planned source", fmt.Sprintf("'%s' Node does not send messages", plan.Source.ValueString()))
			}
		}
	}
	if newSourceMessageType != newTargetMessageType {
		resp.Diagnostics.AddError(
			"Planned source/target MessageType mismatch",
			fmt.Sprintf(
				"%s sends %s, but %s receives %s",
				plan.Source.ValueString(),
				newSourceMessageType,
				plan.Target.ValueString(),
				newTargetMessageType,
			),
		)
	}
}

func (r *EdgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state edgeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadEdge(ctx, r.data.Client, state.Source.ValueString(), state.Target.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading Edge", err.Error())
		return
	} else if echoResp.GetEdge == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		state.Arn = types.StringValue(echoResp.GetEdge.Arn)
		if echoResp.GetEdge.Description != nil {
			state.Description = types.StringValue(*echoResp.GetEdge.Description)
		} else {
			state.Description = types.StringNull()
		}
		if echoResp.GetEdge.KmsKey != nil {
			state.KmsKey = types.StringValue(echoResp.GetEdge.KmsKey.Name)
		} else {
			state.KmsKey = types.StringNull()
		}
		if echoResp.GetEdge.MaxReceiveCount != nil {
			state.MaxReceiveCount = types.Int64Value(int64(*echoResp.GetEdge.MaxReceiveCount))
		}
		state.MessageType = types.StringValue(echoResp.GetEdge.MessageType.Name)
		state.Queue = types.StringValue(echoResp.GetEdge.Queue)
		state.Source = types.StringValue(echoResp.GetEdge.Source.GetName())
		state.Target = types.StringValue(echoResp.GetEdge.Target.GetName())
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EdgeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ARN of the underlying AWS SQS Queue.",
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description.",
				Optional:            true,
			},
			"kmskey": schema.StringAttribute{
				MarkdownDescription: "The name of the KmsKey to use to encrypt the message at rest and in flight. Defaults to the Tenant's KmsKey.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"max_receive_count": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of delivbery tries to the `target`. `0` is the default and will try forever. " +
					"Any positive number will result in that many tries before sending the messagge to the `DeadLetterEmitterNode`.",
				Optional:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{int64validator.AtLeast(0)},
			},
			"message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that will be transmitted.",
			},
			"queue": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL of the underlying AWS SQS queue.",
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "The source Node to transmit messages from.",
				Required:            true,
			},
			"target": schema.StringAttribute{
				MarkdownDescription: "The target Node to transmit messages to.",
				Required:            true,
			},
		},
		MarkdownDescription: "[Edges](https://docs.echo.stream/docs/edges) transmit messages of a single MessageType between Nodes.",
	}
}

func (r *EdgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		description *string
		plan        edgeModel
		state       edgeModel
	)

	// Read Terraform plan and state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}

	if !(state.Source.Equal(plan.Source) && state.Target.Equal(plan.Target)) {
		if echoResp, err := api.MoveEdge(
			ctx,
			r.data.Client,
			state.Source.ValueString(),
			state.Target.ValueString(),
			r.data.Tenant,
			plan.Source.ValueString(),
			plan.Target.ValueString(),
		); err != nil {
			resp.Diagnostics.AddError("Error moving Edge", err.Error())
			return
		} else if echoResp.GetEdge == nil {
			resp.Diagnostics.AddError("Cannot move Edge", fmt.Sprintf("'%s:%s' Edge does not exist", plan.Source.ValueString(), plan.Target.ValueString()))
			return
		} else {
			plan.Arn = types.StringValue(echoResp.GetEdge.Move.Arn)
			if echoResp.GetEdge.Move.Description != nil {
				plan.Description = types.StringValue(*echoResp.GetEdge.Move.Description)
			} else {
				plan.Description = types.StringNull()
			}
			if echoResp.GetEdge.Move.KmsKey != nil {
				plan.KmsKey = types.StringValue(echoResp.GetEdge.Move.KmsKey.Name)
			} else {
				plan.KmsKey = types.StringNull()
			}
			if echoResp.GetEdge.Move.MaxReceiveCount != nil {
				plan.MaxReceiveCount = types.Int64Value(int64(*echoResp.GetEdge.Move.MaxReceiveCount))
			}
			plan.MessageType = types.StringValue(echoResp.GetEdge.Move.MessageType.Name)
			plan.Queue = types.StringValue(echoResp.GetEdge.Move.Queue)
			plan.Source = types.StringValue(echoResp.GetEdge.Move.Source.GetName())
			plan.Target = types.StringValue(echoResp.GetEdge.Move.Target.GetName())
		}
	}

	if echoResp, err := api.UpdateEdge(
		ctx,
		r.data.Client,
		plan.Source.ValueString(),
		plan.Target.ValueString(),
		r.data.Tenant,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error updating Edge", err.Error())
		return
	} else if echoResp.GetEdge == nil {
		resp.Diagnostics.AddError("Cannot update Edge", fmt.Sprintf("'%s:%s' Edge does not exist", plan.Source.ValueString(), plan.Target.ValueString()))
		return
	} else {
		plan.Arn = types.StringValue(echoResp.GetEdge.Update.Arn)
		if echoResp.GetEdge.Update.Description != nil {
			plan.Description = types.StringValue(*echoResp.GetEdge.Update.Description)
		} else {
			plan.Description = types.StringNull()
		}
		if echoResp.GetEdge.Update.KmsKey != nil {
			plan.KmsKey = types.StringValue(echoResp.GetEdge.Update.KmsKey.Name)
		} else {
			plan.KmsKey = types.StringNull()
		}
		if echoResp.GetEdge.Update.MaxReceiveCount != nil {
			plan.MaxReceiveCount = types.Int64Value(int64(*echoResp.GetEdge.Update.MaxReceiveCount))
		}
		plan.MessageType = types.StringValue(echoResp.GetEdge.Update.MessageType.Name)
		plan.Queue = types.StringValue(echoResp.GetEdge.Update.Queue)
		plan.Source = types.StringValue(echoResp.GetEdge.Update.Source.GetName())
		plan.Target = types.StringValue(echoResp.GetEdge.Update.Target.GetName())
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
