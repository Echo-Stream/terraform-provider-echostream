package function

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
		diags               diag.Diagnostics
		readme              *string
		requirements        []string
		return_message_type *string
	)
	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		temp := plan.Readme.ValueString()
		readme = &temp
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.ReturnMessageType.IsNull() || plan.ReturnMessageType.IsUnknown()) {
		temp := plan.ReturnMessageType.ValueString()
		return_message_type = &temp
	}

	echoResp, err := api.CreateProcessorFunction(
		ctx,
		r.data.Client,
		plan.ArgumentMessageType.ValueString(),
		plan.Code.ValueString(),
		plan.Description.ValueString(),
		plan.Name.ValueString(),
		r.data.Tenant,
		readme,
		requirements,
		return_message_type,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating ProcessorFunction", err.Error())
		return
	}

	plan.ArgumentMessageType = types.StringValue(echoResp.CreateProcessorFunction.ArgumentMessageType.Name)
	plan.Code = types.StringValue(echoResp.CreateProcessorFunction.Code)
	plan.InUse = types.BoolValue(echoResp.CreateProcessorFunction.InUse)
	plan.Name = types.StringValue(echoResp.CreateProcessorFunction.Name)
	if echoResp.CreateProcessorFunction.Readme != nil {
		plan.Readme = types.StringValue(*echoResp.CreateProcessorFunction.Readme)
	} else {
		plan.Readme = types.StringNull()
	}
	if len(echoResp.CreateProcessorFunction.Requirements) > 0 {
		elems := []attr.Value{}
		for _, req := range echoResp.CreateProcessorFunction.Requirements {
			elems = append(elems, types.StringValue(req))
		}
		plan.Requirements, diags = types.SetValue(types.StringType, elems)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
	} else {
		plan.Requirements = types.SetNull(types.StringType)
	}
	if echoResp.CreateProcessorFunction.ReturnMessageType != nil {
		plan.ReturnMessageType = types.StringValue(echoResp.CreateProcessorFunction.ReturnMessageType.Name)
	} else {
		plan.ReturnMessageType = types.StringNull()
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

	if _, err := api.DeleteFunction(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ProcessorFunction", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ProcessorFunctionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: resourceProcessorFunctionSchema(),
		MarkdownDescription: "[ProcessorFunctions](https://docs.echo.stream/docs/processor-node#processor-function) provide " +
			"reusable message processing and are used in either a ProcessorNode or a CrossTenantSendingNode.",
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
	if !state.InUse.ValueBool() {
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
		resp.Diagnostics.AddError("Cannot destroy ProcessorFunction", fmt.Sprintf("ProcessorFunction %s is in use and may not be destroyed", state.Name.ValueString()))
	}
}

func (r *ProcessorFunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state processorFunctionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, system, diags := readProcessorFunction(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); diags.HasError() {
		resp.Diagnostics.Append(diags...)
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
		diags        diag.Diagnostics
		readme       *string
		requirements []string
	)
	if !(plan.Code.IsNull() || plan.Code.IsUnknown()) {
		temp := plan.Code.ValueString()
		code = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		temp := plan.Readme.ValueString()
		readme = &temp
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	echoResp, err := api.UpdateFunction(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
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
		plan.ArgumentMessageType = types.StringValue(function.Update.ArgumentMessageType.Name)
		plan.Code = types.StringValue(function.Update.Code)
		plan.Description = types.StringValue(function.Update.Description)
		plan.InUse = types.BoolValue(function.Update.InUse)
		plan.Name = types.StringValue(function.Update.Name)
		if function.Update.Readme != nil {
			plan.Readme = types.StringValue(*function.Update.Readme)
		} else {
			plan.Readme = types.StringNull()
		}
		if len(function.Update.Requirements) > 0 {
			elems := []attr.Value{}
			for _, req := range function.Update.Requirements {
				elems = append(elems, types.StringValue(req))
			}
			plan.Requirements, diags = types.SetValue(types.StringType, elems)
			if diags != nil && diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
		} else {
			plan.Requirements = types.SetNull(types.StringType)
		}
		if function.Update.ReturnMessageType != nil {
			plan.ReturnMessageType = types.StringValue(function.Update.ReturnMessageType.Name)
		} else {
			plan.ReturnMessageType = types.StringNull()
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
