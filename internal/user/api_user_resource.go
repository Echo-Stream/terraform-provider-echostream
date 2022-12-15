package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &ApiUserResource{}
	_ resource.ResourceWithImportState = &ApiUserResource{}
)

type ApiUserResource struct {
	data *common.ProviderData
}

type apiUserModel struct {
	AppsyncEndpoint types.String `tfsdk:"appsync_endpoint"`
	Credentials     types.Object `tfsdk:"credentials"`
	Description     types.String `tfsdk:"description"`
	Role            types.String `tfsdk:"role"`
	Username        types.String `tfsdk:"username"`
}

func (r *ApiUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApiUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiUserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		description *string
		diags       diag.Diagnostics
	)

	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}

	echoResp, err := api.CreateApiUser(
		ctx,
		r.data.Client,
		api.ApiUserRole(plan.Role.ValueString()),
		r.data.Tenant,
		description,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating ApiUser", err.Error())
		return
	}

	plan.AppsyncEndpoint = types.StringValue(echoResp.CreateApiUser.AppsyncEndpoint)
	plan.Credentials, diags = types.ObjectValue(
		common.CognitoCredentialsAttrTypes(),
		common.CognitoCredentialsAttrValues(
			echoResp.CreateApiUser.Credentials.ClientId,
			echoResp.CreateApiUser.Credentials.Password,
			echoResp.CreateApiUser.Credentials.UserPoolId,
			echoResp.CreateApiUser.Credentials.Username,
		),
	)
	if diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	if echoResp.CreateApiUser.Description != nil {
		plan.Description = types.StringValue(*echoResp.CreateApiUser.Description)
	} else {
		plan.Description = types.StringNull()
	}
	plan.Role = types.StringValue(string(echoResp.CreateApiUser.Role))
	plan.Username = types.StringValue(echoResp.CreateApiUser.Username)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApiUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiUserModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteApiUser(ctx, r.data.Client, r.data.Tenant, state.Username.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting ApiUser", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ApiUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("username"), req, resp)
}

func (r *ApiUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_user"
}

func (r *ApiUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		state apiUserModel
	)

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadApiUser(ctx, r.data.Client, r.data.Tenant, state.Username.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error reading ApiUser", err.Error())
		return
	} else if echoResp.GetApiUser == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		state.AppsyncEndpoint = types.StringValue(echoResp.GetApiUser.AppsyncEndpoint)
		state.Credentials, diags = types.ObjectValue(
			common.CognitoCredentialsAttrTypes(),
			common.CognitoCredentialsAttrValues(
				echoResp.GetApiUser.Credentials.ClientId,
				echoResp.GetApiUser.Credentials.Password,
				echoResp.GetApiUser.Credentials.UserPoolId,
				echoResp.GetApiUser.Credentials.Username,
			),
		)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
		if echoResp.GetApiUser.Description != nil {
			state.Description = types.StringValue(*echoResp.GetApiUser.Description)
		} else {
			state.Description = types.StringNull()
		}
		state.Role = types.StringValue(string(echoResp.GetApiUser.Role))
		state.Username = types.StringValue(echoResp.GetApiUser.Username)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApiUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"appsync_endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The EchoStream AppSync Endpoint that this ApiUser must use.",
			},
			"credentials": schema.SingleNestedAttribute{
				Attributes:          common.CognitoCredentialsSchema(),
				Computed:            true,
				MarkdownDescription: "The AWS Cognito Credentials assigned to this ApiUser that must be used when accessing the appsync_endpoint.",
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Human-readble description for this ApiUser.",
				Optional:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("The ApiUser's role. May be on of `%s`, `%s`, or `%s`.", api.ApiUserRoleAdmin, api.ApiUserRoleReadOnly, api.ApiUserRoleUser),
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(api.ApiUserRoleAdmin),
						string(api.ApiUserRoleReadOnly),
						string(api.ApiUserRoleUser),
					),
				},
			},
			"username": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ApiUser's generated username.",
			},
		},
		MarkdownDescription: "ApiUsers are used to programatically interact with your Tenant.",
	}
}

func (r *ApiUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics
		plan  apiUserModel
		state apiUserModel
	)

	// Read Terraform plan and state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}

	roleValue := plan.Role.ValueString()
	echoResp, err := api.UpdateApiUser(
		ctx,
		r.data.Client,
		r.data.Tenant,
		state.Username.ValueString(),
		description,
		(*api.ApiUserRole)(&roleValue),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating ApiUser", err.Error())
		return
	}

	plan.AppsyncEndpoint = types.StringValue(echoResp.GetApiUser.Update.AppsyncEndpoint)
	plan.Credentials, diags = types.ObjectValue(
		common.CognitoCredentialsAttrTypes(),
		common.CognitoCredentialsAttrValues(
			echoResp.GetApiUser.Update.Credentials.ClientId,
			echoResp.GetApiUser.Update.Credentials.Password,
			echoResp.GetApiUser.Update.Credentials.UserPoolId,
			echoResp.GetApiUser.Update.Credentials.Username,
		),
	)
	if diags != nil && diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	if echoResp.GetApiUser.Update.Description != nil {
		plan.Description = types.StringValue(*echoResp.GetApiUser.Update.Description)
	} else {
		plan.Description = types.StringNull()
	}
	plan.Role = types.StringValue(string(echoResp.GetApiUser.Update.Role))
	plan.Username = types.StringValue(echoResp.GetApiUser.Update.Username)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
