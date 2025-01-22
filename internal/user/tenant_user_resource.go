package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &TenantUserResource{}
	_ resource.ResourceWithImportState = &TenantUserResource{}
)

type TenantUserResource struct {
	data *common.ProviderData
}

type tenantUserModel struct {
	Email     types.String `tfsdk:"email"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Role      types.String `tfsdk:"role"`
	Status    types.String `tfsdk:"status"`
}

func (r *TenantUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TenantUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantUserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	echoResp, err := api.CreateTenantUser(
		ctx,
		r.data.Client,
		plan.Email.ValueString(),
		api.UserRole(plan.Role.ValueString()),
		r.data.Tenant,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating TenantUser", err.Error())
		return
	}

	plan.Email = types.StringValue(echoResp.GetTenant.AddUser.Email)
	if echoResp.GetTenant.AddUser.FirstName != nil {
		plan.FirstName = types.StringValue(*echoResp.GetTenant.AddUser.FirstName)
	} else {
		plan.FirstName = types.StringNull()
	}
	if echoResp.GetTenant.AddUser.LastName != nil {
		plan.LastName = types.StringValue(*echoResp.GetTenant.AddUser.LastName)
	} else {
		plan.LastName = types.StringNull()
	}
	plan.Role = types.StringValue(string(echoResp.GetTenant.AddUser.Role))
	plan.Status = types.StringValue(string(echoResp.GetTenant.AddUser.Status))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TenantUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantUserModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteTenantUser(ctx, r.data.Client, state.Email.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting TenantUser", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *TenantUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("email"), req, resp)
}

func (r *TenantUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_user"
}

func (r *TenantUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantUserModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadTenantUser(ctx, r.data.Client, state.Email.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading TenantUser", err.Error())
		return
	} else if echoResp.GetTenantUser == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		state.Email = types.StringValue(echoResp.GetTenantUser.Email)
		if echoResp.GetTenantUser.FirstName != nil {
			state.FirstName = types.StringValue(*echoResp.GetTenantUser.FirstName)
		} else {
			state.FirstName = types.StringNull()
		}
		if echoResp.GetTenantUser.LastName != nil {
			state.LastName = types.StringValue(*echoResp.GetTenantUser.LastName)
		} else {
			state.LastName = types.StringNull()
		}
		state.Role = types.StringValue(string(echoResp.GetTenantUser.Role))
		state.Status = types.StringValue(string(echoResp.GetTenantUser.Status))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TenantUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				MarkdownDescription: "The user's email address.",
				Required:            true,
			},
			"first_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user's first name, if available.",
			},
			"last_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user's last name, if available.",
			},
			"role": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("The ApiUser's role. Must be one of `%s`, `%s`, or `%s`.", api.UserRoleAdmin, api.UserRoleReadOnly, api.UserRoleUser),
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(api.UserRoleAdmin),
						string(api.UserRoleOwner),
						string(api.UserRoleReadOnly),
						string(api.UserRoleUser),
					),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: fmt.Sprintf("The status. If set, must be one of `%s` or `%s`.", api.UserStatusActive, api.UserStatusInactive),
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(api.UserStatusActive),
						string(api.UserStatusInactive),
						string(api.UserStatusInvited),
						string(api.UserStatusPending),
					),
				},
			},
		},
		MarkdownDescription: "[TenantUsers](https://docs.echo.stream/docs/users-1) are used to interact with your Tenant via the UI.",
	}
}

func (r *TenantUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tenantUserModel

	// Read Terraform plan and state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var status *api.UserStatus

	if !(plan.Status.IsNull() || plan.Status.IsUnknown()) {
		if plan.Status.ValueString() == string(api.UserStatusInvited) || plan.Status.ValueString() == string(api.UserStatusPending) {
			resp.Diagnostics.AddAttributeError(
				path.Root("status"),
				"Invalid planned status",
				fmt.Sprintf("status can only be set to %s or %s", api.UserStatusActive, api.UserStatusInactive),
			)
			return
		}
		temp := plan.Status.ValueString()
		status = (*api.UserStatus)(&temp)
	}

	roleValue := plan.Role.ValueString()
	echoResp, err := api.UpdateTenantUser(
		ctx,
		r.data.Client,
		plan.Email.ValueString(),
		r.data.Tenant,
		(*api.UserRole)(&roleValue),
		status,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating TenantUser", err.Error())
		return
	}

	plan.Email = types.StringValue(echoResp.GetTenantUser.Update.Email)
	if echoResp.GetTenantUser.Update.FirstName != nil {
		plan.FirstName = types.StringValue(*echoResp.GetTenantUser.Update.FirstName)
	} else {
		plan.FirstName = types.StringNull()
	}
	if echoResp.GetTenantUser.Update.LastName != nil {
		plan.LastName = types.StringValue(*echoResp.GetTenantUser.Update.LastName)
	} else {
		plan.LastName = types.StringNull()
	}
	plan.Role = types.StringValue(string(echoResp.GetTenantUser.Update.Role))
	plan.Status = types.StringValue(string(echoResp.GetTenantUser.Update.Status))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
