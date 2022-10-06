package function

import (
	"context"
	"fmt"

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
	_ resource.ResourceWithImportState = &BitmapperFunctionResource{}
	_ resource.ResourceWithModifyPlan  = &BitmapperFunctionResource{}
)

// BitmapperFunctionResource defines the resource implementation.
type BitmapperFunctionResource struct {
	data *common.ProviderData
}

func (r *BitmapperFunctionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BitmapperFunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bitmapperFunctionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		readme       *string
		requirements []string
	)
	if !plan.Readme.IsNull() {
		readme = &plan.Readme.Value
	}
	if !plan.Requirements.IsNull() {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	echoResp, err := api.CreateBitmapperFunction(
		ctx,
		r.data.Client,
		plan.ArgumentMessageType.Value,
		plan.Code.Value,
		plan.Description.Value,
		plan.Name.Value,
		r.data.Tenant,
		readme,
		requirements,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating BitmapperFunction", err.Error())
	}

	plan.ArgumentMessageType = types.String{Value: echoResp.CreateBitmapperFunction.ArgumentMessageType.Name}
	plan.Code = types.String{Value: echoResp.CreateBitmapperFunction.Code}
	plan.InUse = types.Bool{Value: echoResp.CreateBitmapperFunction.InUse}
	plan.Name = types.String{Value: echoResp.CreateBitmapperFunction.Name}
	if echoResp.CreateBitmapperFunction.Readme != nil {
		plan.Readme = types.String{Value: *echoResp.CreateBitmapperFunction.Readme}
	} else {
		plan.Readme = types.String{Null: true}
	}
	if len(echoResp.CreateBitmapperFunction.Requirements) > 0 {
		plan.Requirements = types.Set{ElemType: types.StringType}
		for _, req := range echoResp.CreateBitmapperFunction.Requirements {
			plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
		}
	} else {
		plan.Requirements = types.Set{ElemType: types.StringType, Null: true}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BitmapperFunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bitmapperFunctionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteFunction(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting BitmapperFunction", err.Error())
	}
}

func (r *BitmapperFunctionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:          resourceBitmapperFunctionSchema(),
		Description:         "BitmapperFunctions provide reusable message bitmapping",
		MarkdownDescription: "BitmapperFunctions provide reusable message bitmapping",
	}, nil
}

func (r *BitmapperFunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *BitmapperFunctionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bitmapper_function"
}

func (r *BitmapperFunctionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state bitmapperFunctionModel

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
		var plan bitmapperFunctionModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		prevent_destroy = !(plan.Name.Equal(state.Name) && plan.ArgumentMessageType.Equal(state.ArgumentMessageType))
	}

	if prevent_destroy {
		resp.Diagnostics.AddError("Cannot destroy BitmapperFunction", fmt.Sprintf("BitmapperFunction %s is in use and may not be destroyed", state.Name.Value))
	}
}

func (r *BitmapperFunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bitmapperFunctionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if system, err := readBitmapperFunction(ctx, r.data.Client, state.Name.Value, r.data.Tenant, &state); err != nil {
		resp.Diagnostics.AddError("Error reading BitmapperFunction", err.Error())
		return
	} else if system {
		resp.Diagnostics.AddError("Invalid BitmapperFunction", "Cannot create resource for system BitmapperFunction")
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BitmapperFunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *bitmapperFunctionModel

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
	if !plan.Code.IsNull() {
		code = &plan.Code.Value
	}
	if !plan.Description.IsNull() {
		description = &plan.Description.Value
	}
	if !plan.Readme.IsNull() {
		readme = &plan.Readme.Value
	}
	if !plan.Requirements.IsNull() {
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
		resp.Diagnostics.AddError("Error updating BitmapperFunction", err.Error())
	}

	switch function := (*echoResp.GetFunction).(type) {
	case *api.UpdateFunctionGetFunctionBitmapperFunction:
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
		if len(function.Update.Requirements) > 0 {
			plan.Requirements = types.Set{ElemType: types.StringType}
			for _, req := range function.Update.Requirements {
				plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
			}
		} else {
			plan.Requirements = types.Set{ElemType: types.StringType, Null: true}
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
