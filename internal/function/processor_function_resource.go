package function

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
	_ resource.ResourceWithImportState = &ProcessorFunctionResource{}
	_ resource.ResourceWithModifyPlan  = &ProcessorFunctionResource{}
)

// BitmapperFunctionResource defines the resource implementation.
type ProcessorFunctionResource struct {
	data *common.ProviderData
}

func (r *ProcessorFunctionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProcessorFunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan processorFunctionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		readme              *string
		requirements        []string
		return_message_type *string
	)
	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		readme = &plan.Readme.Value
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.ReturnMessageType.IsNull() || plan.ReturnMessageType.IsUnknown()) {
		return_message_type = &plan.ReturnMessageType.Value
	}

	echoResp, err := api.CreateProcessorFunction(
		ctx,
		r.data.Client,
		plan.ArgumentMessageType.Value,
		plan.Code.Value,
		plan.Description.Value,
		plan.Name.Value,
		r.data.Tenant,
		readme,
		requirements,
		return_message_type,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating ProcessorFunction", err.Error())
		return
	}

	plan.ArgumentMessageType = types.String{Value: echoResp.CreateProcessorFunction.ArgumentMessageType.Name}
	plan.Code = types.String{Value: echoResp.CreateProcessorFunction.Code}
	plan.InUse = types.Bool{Value: echoResp.CreateProcessorFunction.InUse}
	plan.Name = types.String{Value: echoResp.CreateProcessorFunction.Name}
	if echoResp.CreateProcessorFunction.Readme != nil {
		plan.Readme = types.String{Value: *echoResp.CreateProcessorFunction.Readme}
	} else {
		plan.Readme = types.String{Null: true}
	}
	plan.Requirements = types.Set{ElemType: types.StringType}
	if len(echoResp.CreateProcessorFunction.Requirements) > 0 {
		for _, req := range echoResp.CreateProcessorFunction.Requirements {
			plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
		}
	} else {
		plan.Requirements.Null = true
	}
	if echoResp.CreateProcessorFunction.ReturnMessageType != nil {
		plan.ReturnMessageType = types.String{Value: echoResp.CreateProcessorFunction.ReturnMessageType.Name}
	} else {
		plan.ReturnMessageType = types.String{Null: true}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProcessorFunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state processorFunctionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteFunction(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ProcessorFunction", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ProcessorFunctionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:          resourceProcessorFunctionSchema(),
		Description:         "ProcessorFunctions provide reusable message processing",
		MarkdownDescription: "ProcessorFunctions provide reusable message processing",
	}, nil
}

func (r *ProcessorFunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ProcessorFunctionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_processor_function"
}

func (r *ProcessorFunctionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state processorFunctionModel

	// If the entire state is null, resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If the MessageType is not in use it can be destroyed at will.
	if !state.InUse.Value {
		return
	}

	// If the entire plan is null, the resource is planned for destruction.
	prevent_destroy := req.Plan.Raw.IsNull()
	if !prevent_destroy {
		var plan processorFunctionModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		prevent_destroy = !(plan.Name.Equal(state.Name) && plan.ArgumentMessageType.Equal(state.ArgumentMessageType) && plan.ReturnMessageType.Equal(state.ReturnMessageType))
	}

	if prevent_destroy {
		resp.Diagnostics.AddError("Cannot destroy ProcessorFunction", fmt.Sprintf("ProcessorFunction %s is in use and may not be destroyed", state.Name.Value))
	}
}

func (r *ProcessorFunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state processorFunctionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, system, err := readProcessorFunction(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ProcessorFunction", err.Error())
		return
	} else if system {
		resp.Diagnostics.AddError("Invalid ProcessorFunction", "Cannot create resource for system ProcessorFunction")
		return
	} else if data == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		state = *data
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProcessorFunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan processorFunctionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		code         *string
		description  *string
		readme       *string
		requirements []string
	)
	if !(plan.Code.IsNull() || plan.Code.IsUnknown()) {
		code = &plan.Code.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		readme = &plan.Readme.Value
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	echoResp, err := api.UpdateFunction(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		code,
		description,
		readme,
		requirements,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating ProcessorFunction", err.Error())
		return
	}

	switch function := (*echoResp.GetFunction).(type) {
	case *api.UpdateFunctionGetFunctionProcessorFunction:
		plan.ArgumentMessageType = types.String{Value: function.Update.ArgumentMessageType.Name}
		plan.Code = types.String{Value: function.Update.Code}
		plan.Description = types.String{Value: function.Update.Description}
		plan.InUse = types.Bool{Value: function.Update.InUse}
		plan.Name = types.String{Value: function.Update.Name}
		if function.Update.Readme != nil {
			plan.Readme = types.String{Value: *function.Update.Readme}
		} else {
			plan.Readme = types.String{Null: true}
		}
		plan.Requirements = types.Set{ElemType: types.StringType}
		if len(function.Update.Requirements) > 0 {
			for _, req := range function.Update.Requirements {
				plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
			}
		} else {
			plan.Requirements.Null = true
		}
		if function.Update.ReturnMessageType != nil {
			plan.ReturnMessageType = types.String{Value: function.Update.ReturnMessageType.Name}
		} else {
			plan.ReturnMessageType = types.String{Null: true}
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
