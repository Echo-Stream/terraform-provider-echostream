package app

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
	_ resource.ResourceWithImportState = &CrossTenantReceivingAppResource{}
	_ resource.ResourceWithModifyPlan  = &CrossTenantReceivingAppResource{}
)

// CrossTenantReceivingAppResource defines the resource implementation.
type CrossTenantReceivingAppResource struct {
	data *common.ProviderData
}

func (r *CrossTenantReceivingAppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CrossTenantReceivingAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan crossTenantReceivingAppModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.CreateCrossTenantReceivingApp(
		ctx,
		r.data.Client,
		plan.Name.Value,
		plan.SendingTenant.Value,
		r.data.Tenant,
	); err != nil {
		resp.Diagnostics.AddError("Error creating CrossTenantReceivingApp", err.Error())
		return
	} else {
		plan.Description = types.String{Value: *echoResp.CreateCrossTenantReceivingApp.Description}
		plan.Name = types.String{Value: echoResp.CreateCrossTenantReceivingApp.Name}
		plan.SendingApp = types.String{Null: true}
		plan.SendingTenant = types.String{Value: echoResp.CreateCrossTenantReceivingApp.SendingTenant}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CrossTenantReceivingAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state crossTenantReceivingAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteApp(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting CrossTenentReceivingApp", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *CrossTenantReceivingAppResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:          crossTenantReceivingSchema(),
		Description:         "CrossTenantReceivingApps provide a way to receive messages from other EchoStream Tenants",
		MarkdownDescription: "CrossTenantReceivingApps provide a way to receive messages from other EchoStream Tenants",
	}, nil
}

func (r *CrossTenantReceivingAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *CrossTenantReceivingAppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cross_tenant_receiving_app"
}

func (r *CrossTenantReceivingAppResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state crossTenantReceivingAppModel

	// If the entire state is null, resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If the CrossTenantReceivingApp is not in use it can be destroyed at will.
	if state.SendingApp.IsNull() {
		return
	}

	// If the entire plan is null, the resource is planned for destruction.
	if req.Plan.Raw.IsNull() {
		resp.Diagnostics.AddWarning(
			"Destroying connected CrossTenantReceivingApp",
			fmt.Sprintf("This will terminate the connection with %s permanently!!", state.SendingApp.Value),
		)
	} else {
		var plan crossTenantReceivingAppModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		if !(plan.Name.Equal(state.Name) && plan.SendingTenant.Equal(state.SendingTenant)) {
			resp.Diagnostics.AddWarning(
				"Replacing connected CrossTenantReceivingApp",
				fmt.Sprintf("This will terminate the connection with %s permanently!!", state.SendingApp.Value),
			)
		}
	}
}

func (r *CrossTenantReceivingAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state crossTenantReceivingAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadApp(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossTenantReceivingApp", err.Error())
		return
	} else if echoResp.GetApp != nil {
		switch app := (*echoResp.GetApp).(type) {
		case *api.ReadAppGetAppCrossTenantReceivingApp:
			state.Description = types.String{Value: *app.Description}
			state.Name = types.String{Value: app.Name}
			if app.SendingApp != nil {
				state.SendingApp = types.String{Value: *app.SendingApp}
			} else {
				state.SendingApp = types.String{Null: true}
			}
			state.SendingTenant = types.String{Value: app.SendingTenant}
		default:
			resp.Diagnostics.AddError(
				"Incorrect App type",
				fmt.Sprintf("'%s' is incorrect App type", state.Name.String()),
			)
			return
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CrossTenantReceivingAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan crossTenantReceivingAppModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string

	if !plan.Description.IsNull() {
		description = &plan.Description.Value
	}

	echoResp, err := api.UpdateCrossTenantApp(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		description,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating CrossTenantReceivingApp", err.Error())
		return
	}

	switch app := (*echoResp.GetApp).(type) {
	case *api.UpdateCrossTenantAppGetAppCrossTenantReceivingApp:
		plan.Description = types.String{Value: *app.Update.Description}
		plan.Name = types.String{Value: app.Update.Name}
		if app.Update.SendingApp != nil {
			plan.SendingApp = types.String{Value: *app.Update.SendingApp}
		} else {
			plan.SendingApp = types.String{Null: true}
		}
		plan.SendingTenant = types.String{Value: app.Update.SendingTenant}
	default:
		resp.Diagnostics.AddError(
			"Incorrect App type",
			fmt.Sprintf("'%s' is incorrect App type", plan.Name.String()),
		)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
