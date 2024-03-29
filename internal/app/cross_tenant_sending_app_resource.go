package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigure   = &CrossTenantSendingAppResource{}
	_ resource.ResourceWithImportState = &CrossTenantSendingAppResource{}
	_ resource.ResourceWithModifyPlan  = &CrossTenantSendingAppResource{}
)

// CrossTenantSendingAppResource defines the resource implementation.
type CrossTenantSendingAppResource struct {
	data *common.ProviderData
}

type crossTenantSendingAppModel struct {
	Description     types.String `tfsdk:"description"`
	Name            types.String `tfsdk:"name"`
	ReceivingApp    types.String `tfsdk:"receiving_app"`
	ReceivingTenant types.String `tfsdk:"receiving_tenant"`
}

func (r *CrossTenantSendingAppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CrossTenantSendingAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan crossTenantSendingAppModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.CreateCrossTenantSendingApp(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		plan.ReceivingApp.ValueString(),
		plan.ReceivingTenant.ValueString(),
		r.data.Tenant,
	); err != nil {
		resp.Diagnostics.AddError("Error creating CrossTenantSendingApp", err.Error())
		return
	} else {
		if echoResp.CreateCrossTenantSendingApp.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateCrossTenantSendingApp.Description)
		} else {
			plan.Description = types.StringNull()
		}
		plan.Name = types.StringValue(echoResp.CreateCrossTenantSendingApp.Name)
		plan.ReceivingApp = types.StringValue(echoResp.CreateCrossTenantSendingApp.ReceivingApp)
		plan.ReceivingTenant = types.StringValue(echoResp.CreateCrossTenantSendingApp.ReceivingTenant)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CrossTenantSendingAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state crossTenantSendingAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteApp(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting CrossTenantSendingApp", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *CrossTenantSendingAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *CrossTenantSendingAppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cross_tenant_sending_app"
}

func (r *CrossTenantSendingAppResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state crossTenantSendingAppModel

	// If the entire state is null, resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If the entire plan is null, the resource is planned for destruction.
	if req.Plan.Raw.IsNull() {
		resp.Diagnostics.AddWarning(
			"Destroying connected CrossTenantSendingApp",
			fmt.Sprintf("This will terminate the connection with %s permanently!!", state.ReceivingApp.ValueString()),
		)
	} else {
		var plan crossTenantSendingAppModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		if !(plan.Name.Equal(state.Name) &&
			plan.ReceivingTenant.Equal(state.ReceivingTenant) &&
			plan.ReceivingApp.Equal(state.ReceivingApp)) {
			resp.Diagnostics.AddWarning(
				"Replacing connected CrossTenantSendingApp",
				fmt.Sprintf("This will terminate the connection with %s permanently!!", state.ReceivingApp.ValueString()),
			)
		}
	}
}

func (r *CrossTenantSendingAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state crossTenantSendingAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadApp(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossTenantSendingApp", err.Error())
		return
	} else if echoResp.GetApp != nil {
		switch app := (*echoResp.GetApp).(type) {
		case *api.ReadAppGetAppCrossTenantSendingApp:
			state.Description = types.StringValue(*app.Description)
			state.Name = types.StringValue(app.Name)
			state.ReceivingApp = types.StringValue(app.ReceivingApp)
			state.ReceivingTenant = types.StringValue(app.ReceivingTenant)
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

func (r *CrossTenantSendingAppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := appResourceAttributes()
	maps.Copy(
		attributes,
		map[string]schema.Attribute{
			"receiving_app": schema.StringAttribute{
				MarkdownDescription: "The CrossTenantReceivingApp in the `receiving_tenant`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
			},
			"receiving_tenant": schema.StringAttribute{
				MarkdownDescription: "The EchoStream Tenant that you will send message to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
			},
		},
	)
	description := attributes["description"].(schema.StringAttribute)
	description.Computed = true
	attributes["description"] = description
	resp.Schema = schema.Schema{
		Attributes:          attributes,
		MarkdownDescription: "[CrossTenantSendingApps](https://docs.echo.stream/docs/cross-tenant-app) provide a way to send messages to another EchoStream Tenant.",
	}
}

func (r *CrossTenantSendingAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan crossTenantSendingAppModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string

	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}

	echoResp, err := api.UpdateCrossTenantApp(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		r.data.Tenant,
		description,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating CrossTenantSendingApp", err.Error())
		return
	}

	switch app := (*echoResp.GetApp).(type) {
	case *api.UpdateCrossTenantAppGetAppCrossTenantSendingApp:
		plan.Description = types.StringValue(*app.Update.Description)
		plan.Name = types.StringValue(app.Update.Name)
		plan.ReceivingApp = types.StringValue(app.Update.ReceivingApp)
		plan.ReceivingTenant = types.StringValue(app.Update.ReceivingTenant)
	default:
		resp.Diagnostics.AddError(
			"Incorrect App type",
			fmt.Sprintf("'%s' is incorrect App type", plan.Name.String()),
		)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
