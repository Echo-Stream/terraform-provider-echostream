package edge

import (
	"context"
	"fmt"
	"reflect"
	"strings"
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
	if !plan.Description.IsNull() {
		description = &plan.Description.Value
	}
	if !plan.KmsKey.IsNull() {
		kmsKey = &plan.KmsKey.Value
	}
	if !plan.MaxReceiveCount.IsNull() {
		mrc := int(plan.MaxReceiveCount.Value)
		maxReceiveCount = &mrc
	}

	if echoResp, err := api.CreateEdge(
		ctx,
		r.data.Client,
		plan.Source.Value,
		plan.Target.Value,
		r.data.Tenant,
		description,
		kmsKey,
		maxReceiveCount,
	); err != nil {
		resp.Diagnostics.AddError("Error creating Edge", err.Error())
		return
	} else {
		if echoResp.CreateEdge.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateEdge.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		if echoResp.CreateEdge.KmsKey != nil {
			plan.KmsKey = types.String{Value: echoResp.CreateEdge.KmsKey.Name}
		} else {
			plan.KmsKey = types.String{Null: true}
		}
		if echoResp.CreateEdge.MaxReceiveCount != nil {
			plan.MaxReceiveCount = types.Int64{Value: int64(*echoResp.CreateEdge.MaxReceiveCount)}
		}
		plan.MessageType = types.String{Value: echoResp.CreateEdge.MessageType.Name}
		plan.Queue = types.String{Value: echoResp.CreateEdge.Queue}
		plan.Source = types.String{Value: echoResp.CreateEdge.Source.GetName()}
		plan.Target = types.String{Value: echoResp.CreateEdge.Target.GetName()}
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

	if _, err := api.DeleteEdge(ctx, r.data.Client, state.Source.Value, state.Target.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting Edge", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *EdgeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"description": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
			},
			"kmskey": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Type:                types.StringType,
			},
			"max_receive_count": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Type:                types.Int64Type,
			},
			"message_type": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
			"queue": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
			"source": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
			},
			"target": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
			},
		},
		Description:         "",
		MarkdownDescription: "",
	}, nil
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

	newSourceMessageType := state.MessageType.Value
	newTargetMessageType := state.MessageType.Value

	if !state.Source.Equal(plan.Source) {
		if echoResp, err := api.ReadNodeMessageTypes(ctx, r.data.Client, plan.Source.Value, r.data.Tenant); err != nil {
			resp.Diagnostics.AddError("Error reading planned source", err.Error())
			return
		} else if echoResp.GetNode == nil {
			resp.Diagnostics.AddAttributeError(path.Root("source"), "Cannot find Node", fmt.Sprintf("'%s' Node does not exist", plan.Source.Value))
			return
		} else {
			node := reflect.Indirect(reflect.ValueOf(*echoResp.GetNode))
			if smt := reflect.Indirect(node.FieldByName("SendMessageType")); !smt.IsZero() {
				if name := smt.FieldByName("Name").String(); name != state.MessageType.Value {
					newSourceMessageType = name
					resp.RequiresReplace = append(resp.RequiresReplace, path.Root("source"))
				}
			} else {
				resp.Diagnostics.AddAttributeError(path.Root("source"), "Invalid planned source", fmt.Sprintf("'%s' Node does not send messages", plan.Source.Value))
			}
		}
	}
	if !state.Target.Equal(plan.Target) {
		if echoResp, err := api.ReadNodeMessageTypes(ctx, r.data.Client, plan.Target.Value, r.data.Tenant); err != nil {
			resp.Diagnostics.AddError("Error reading planned target", err.Error())
			return
		} else if echoResp.GetNode == nil {
			resp.Diagnostics.AddAttributeError(path.Root("target"), "Cannot find Node", fmt.Sprintf("'%s' Node does not exist", plan.Target.Value))
			return
		} else {
			node := reflect.Indirect(reflect.ValueOf(*echoResp.GetNode))
			if smt := reflect.Indirect(node.FieldByName("ReceiveMessageType")); !smt.IsZero() {
				if name := smt.FieldByName("Name").String(); name != state.MessageType.Value {
					newTargetMessageType = name
					resp.RequiresReplace = append(resp.RequiresReplace, path.Root("target"))
				}
			} else {
				resp.Diagnostics.AddAttributeError(path.Root("source"), "Invalid planned source", fmt.Sprintf("'%s' Node does not send messages", plan.Source.Value))
			}
		}
	}
	if newSourceMessageType != newTargetMessageType {
		resp.Diagnostics.AddError(
			"Planned source/target MessageType mismatch",
			fmt.Sprintf(
				"%s sends %s, but %s receives %s",
				plan.Source.Value,
				newSourceMessageType,
				plan.Target.Value,
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

	if echoResp, err := api.ReadEdge(ctx, r.data.Client, state.Source.Value, state.Target.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading Edge", err.Error())
		return
	} else if echoResp.GetEdge == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		if echoResp.GetEdge.Description != nil {
			state.Description = types.String{Value: *echoResp.GetEdge.Description}
		} else {
			state.Description = types.String{Null: true}
		}
		if echoResp.GetEdge.KmsKey != nil {
			state.KmsKey = types.String{Value: echoResp.GetEdge.KmsKey.Name}
		} else {
			state.KmsKey = types.String{Null: true}
		}
		if echoResp.GetEdge.MaxReceiveCount != nil {
			state.MaxReceiveCount = types.Int64{Value: int64(*echoResp.GetEdge.MaxReceiveCount)}
		}
		state.MessageType = types.String{Value: echoResp.GetEdge.MessageType.Name}
		state.Queue = types.String{Value: echoResp.GetEdge.Queue}
		state.Source = types.String{Value: echoResp.GetEdge.Source.GetName()}
		state.Target = types.String{Value: echoResp.GetEdge.Target.GetName()}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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

	if !plan.Description.IsNull() {
		description = &plan.Description.Value
	}

	if !(state.Source.Equal(plan.Source) && state.Target.Equal(plan.Target)) {
		if echoResp, err := api.MoveEdge(
			ctx,
			r.data.Client,
			state.Source.Value,
			state.Target.Value,
			r.data.Tenant,
			plan.Source.Value,
			plan.Target.Value,
		); err != nil {
			resp.Diagnostics.AddError("Error moving Edge", err.Error())
			return
		} else if echoResp.GetEdge == nil {
			resp.Diagnostics.AddError("Cannot move Edge", fmt.Sprintf("'%s:%s' Edge does not exist", plan.Source.Value, plan.Target.Value))
			return
		} else {
			if echoResp.GetEdge.Move.Description != nil {
				plan.Description = types.String{Value: *echoResp.GetEdge.Move.Description}
			} else {
				plan.Description = types.String{Null: true}
			}
			if echoResp.GetEdge.Move.KmsKey != nil {
				plan.KmsKey = types.String{Value: echoResp.GetEdge.Move.KmsKey.Name}
			} else {
				plan.KmsKey = types.String{Null: true}
			}
			if echoResp.GetEdge.Move.MaxReceiveCount != nil {
				plan.MaxReceiveCount = types.Int64{Value: int64(*echoResp.GetEdge.Move.MaxReceiveCount)}
			}
			plan.MessageType = types.String{Value: echoResp.GetEdge.Move.MessageType.Name}
			plan.Queue = types.String{Value: echoResp.GetEdge.Move.Queue}
			plan.Source = types.String{Value: echoResp.GetEdge.Move.Source.GetName()}
			plan.Target = types.String{Value: echoResp.GetEdge.Move.Target.GetName()}
		}
	}

	if echoResp, err := api.UpdateEdge(
		ctx,
		r.data.Client,
		plan.Source.Value,
		plan.Target.Value,
		r.data.Tenant,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error updating Edge", err.Error())
		return
	} else if echoResp.GetEdge == nil {
		resp.Diagnostics.AddError("Cannot update Edge", fmt.Sprintf("'%s:%s' Edge does not exist", plan.Source.Value, plan.Target.Value))
		return
	} else {
		if echoResp.GetEdge.Update.Description != nil {
			plan.Description = types.String{Value: *echoResp.GetEdge.Update.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		if echoResp.GetEdge.Update.KmsKey != nil {
			plan.KmsKey = types.String{Value: echoResp.GetEdge.Update.KmsKey.Name}
		} else {
			plan.KmsKey = types.String{Null: true}
		}
		if echoResp.GetEdge.Update.MaxReceiveCount != nil {
			plan.MaxReceiveCount = types.Int64{Value: int64(*echoResp.GetEdge.Update.MaxReceiveCount)}
		}
		plan.MessageType = types.String{Value: echoResp.GetEdge.Update.MessageType.Name}
		plan.Queue = types.String{Value: echoResp.GetEdge.Update.Queue}
		plan.Source = types.String{Value: echoResp.GetEdge.Update.Source.GetName()}
		plan.Target = types.String{Value: echoResp.GetEdge.Update.Target.GetName()}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
