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
	_ resource.ResourceWithImportState = &ApiAuthenticatorFunctionResource{}
	_ resource.ResourceWithModifyPlan  = &ApiAuthenticatorFunctionResource{}
)

// ApiAuthenticatorFunctionResource defines the resource implementation.
type ApiAuthenticatorFunctionResource struct {
	data *common.ProviderData
}

func (r *ApiAuthenticatorFunctionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApiAuthenticatorFunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan functionModel

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

	echoResp, err := api.CreateApiAuthenticatorFunction(
		ctx,
		r.data.Client,
		plan.Code.Value,
		plan.Description.Value,
		plan.Name.Value,
		r.data.Tenant,
		readme,
		requirements,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating ApiAuthenticatorFunction", err.Error())
		return
	}

	plan.Code = types.String{Value: echoResp.CreateApiAuthenticatorFunction.Code}
	plan.InUse = types.Bool{Value: echoResp.CreateApiAuthenticatorFunction.InUse}
	plan.Name = types.String{Value: echoResp.CreateApiAuthenticatorFunction.Name}
	if echoResp.CreateApiAuthenticatorFunction.Readme != nil {
		plan.Readme = types.String{Value: *echoResp.CreateApiAuthenticatorFunction.Readme}
	} else {
		plan.Readme = types.String{Null: true}
	}
	if len(echoResp.CreateApiAuthenticatorFunction.Requirements) > 0 {
		plan.Requirements = types.Set{ElemType: types.StringType}
		for _, req := range echoResp.CreateApiAuthenticatorFunction.Requirements {
			plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
		}
	} else {
		plan.Requirements = types.Set{ElemType: types.StringType, Null: true}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApiAuthenticatorFunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state functionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteFunction(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ApiAuthenticatorFunction", err.Error())
		return
	}
}

func (r *ApiAuthenticatorFunctionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:          resourceFunctionSchema(),
		Description:         "ApiAuthenticatorFunctions provide reusable api authentication for various nodes",
		MarkdownDescription: "ApiAuthenticatorFunctions provide reusable api authentication for various nodes",
	}, nil
}

func (r *ApiAuthenticatorFunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ApiAuthenticatorFunctionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_authenticator_function"
}

func (r *ApiAuthenticatorFunctionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state functionModel

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
		var plan functionModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		prevent_destroy = !plan.Name.Equal(state.Name)
	}

	if prevent_destroy {
		resp.Diagnostics.AddError("Cannot destroy ApiAuthenticatorFunction", fmt.Sprintf("ApiAuthenticatorFunction %s is in use and may not be destroyed", state.Name.Value))
	}
}

func (r *ApiAuthenticatorFunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state functionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if system, err := readFunction(ctx, r.data.Client, state.Name.Value, r.data.Tenant, &state); err != nil {
		resp.Diagnostics.AddError("Error reading ApiAuthenticatorFunction", err.Error())
		return
	} else if system {
		resp.Diagnostics.AddError("Invalid ApiAuthenticatorFunction", "Cannot create resource for system ApiAuthenticatorFunction")
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApiAuthenticatorFunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *functionModel

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
		resp.Diagnostics.AddError("Error updating ApiAuthenticatorFunction", err.Error())
		return
	}

	switch function := (*echoResp.GetFunction).(type) {
	case *api.UpdateFunctionGetFunctionApiAuthenticatorFunction:
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
