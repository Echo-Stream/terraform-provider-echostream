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
	_ resource.ResourceWithConfigure   = &CrossTenantReceivingAppResource{}
	_ resource.ResourceWithImportState = &CrossTenantReceivingAppResource{}
	_ resource.ResourceWithModifyPlan  = &CrossTenantReceivingAppResource{}
)

// CrossTenantReceivingAppResource defines the resource implementation.
type CrossTenantReceivingAppResource struct {
	data *common.ProviderData
}

type crossTenantReceivingAppModel struct {
	Description   types.String `tfsdk:"description"`
	Name          types.String `tfsdk:"name"`
	SendingApp    types.String `tfsdk:"sending_app"`
	SendingTenant types.String `tfsdk:"sending_tenant"`
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
		plan.Name.ValueString(),
		plan.SendingTenant.ValueString(),
		r.data.Tenant,
	); err != nil {
		resp.Diagnostics.AddError("Error creating CrossTenantReceivingApp", err.Error())
		return
	} else {
		plan.Description = types.StringValue(*echoResp.CreateCrossTenantReceivingApp.Description)
		plan.Name = types.StringValue(echoResp.CreateCrossTenantReceivingApp.Name)
		plan.SendingApp = types.StringNull()
		plan.SendingTenant = types.StringValue(echoResp.CreateCrossTenantReceivingApp.SendingTenant)
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

	if _, err := api.DeleteApp(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting CrossTenentReceivingApp", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
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
			fmt.Sprintf("This will terminate the connection with %s permanently!!", state.SendingApp.ValueString()),
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
				fmt.Sprintf("This will terminate the connection with %s permanently!!", state.SendingApp.ValueString()),
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

	if echoResp, err := api.ReadApp(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossTenantReceivingApp", err.Error())
		return
	} else if echoResp.GetApp != nil {
		switch app := (*echoResp.GetApp).(type) {
		case *api.ReadAppGetAppCrossTenantReceivingApp:
			state.Description = types.StringValue(*app.Description)
			state.Name = types.StringValue(app.Name)
			if app.SendingApp != nil {
				state.SendingApp = types.StringValue(*app.SendingApp)
			} else {
				state.SendingApp = types.StringNull()
			}
			state.SendingTenant = types.StringValue(app.SendingTenant)
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

func (r *CrossTenantReceivingAppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := appResourceAttributes()
	maps.Copy(
		attributes,
		map[string]schema.Attribute{
			"sending_app": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The CrossTenantSendingApp in the sending Tenant - this will be filled in once the other Tenant creates their CrossTenantSendingApp.",
			},
			"sending_tenant": schema.StringAttribute{
				MarkdownDescription: "The EchoStream Tenant that will be sending messages to this CrossTenantReceivingApp.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
			},
		},
	)
	resp.Schema = schema.Schema{
		Attributes:          attributes,
		MarkdownDescription: "[CrossTenantReceivingApps](https://docs.echo.stream/docs/cross-tenant-app) provide a way to receive messages from other EchoStream Tenants.",
	}
}

func (r *CrossTenantReceivingAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan crossTenantReceivingAppModel

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
		resp.Diagnostics.AddError("Error updating CrossTenantReceivingApp", err.Error())
		return
	}

	switch app := (*echoResp.GetApp).(type) {
	case *api.UpdateCrossTenantAppGetAppCrossTenantReceivingApp:
		plan.Description = types.StringValue(*app.Update.Description)
		plan.Name = types.StringValue(app.Update.Name)
		if app.Update.SendingApp != nil {
			plan.SendingApp = types.StringValue(*app.Update.SendingApp)
		} else {
			plan.SendingApp = types.StringNull()
		}
		plan.SendingTenant = types.StringValue(app.Update.SendingTenant)
	default:
		resp.Diagnostics.AddError(
			"Incorrect App type",
			fmt.Sprintf("'%s' is incorrect App type", plan.Name.String()),
		)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
