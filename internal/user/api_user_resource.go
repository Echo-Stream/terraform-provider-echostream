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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithImportState = &ApiUserResource{}

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

	var description *string

	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}

	echoResp, err := api.CreateApiUser(
		ctx,
		r.data.Client,
		api.ApiUserRole(plan.Role.Value),
		r.data.Tenant,
		description,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating ApiUser", err.Error())
		return
	}

	plan.AppsyncEndpoint = types.String{Value: echoResp.CreateApiUser.AppsyncEndpoint}
	plan.Credentials = types.Object{
		Attrs: common.CognitoCredentialsAttrValues(
			echoResp.CreateApiUser.Credentials.ClientId,
			echoResp.CreateApiUser.Credentials.Password,
			echoResp.CreateApiUser.Credentials.UserPoolId,
			echoResp.CreateApiUser.Credentials.Username,
		),
		AttrTypes: common.CognitoCredentialsAttrTypes(),
	}
	if echoResp.CreateApiUser.Description != nil {
		plan.Description = types.String{Value: *echoResp.CreateApiUser.Description}
	} else {
		plan.Description = types.String{Null: true}
	}
	plan.Role = types.String{Value: string(echoResp.CreateApiUser.Role)}
	plan.Username = types.String{Value: echoResp.CreateApiUser.Username}

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

	if _, err := api.DeleteApiUser(ctx, r.data.Client, r.data.Tenant, state.Username.Value); err != nil {
		resp.Diagnostics.AddError("Error deleting ApiUser", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ApiUserResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"appsync_endpoint": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
			"credentials": {
				Attributes:          tfsdk.SingleNestedAttributes(common.CognitoCredentialsSchema()),
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
			},
			"description": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
			},
			"role": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf(
						string(api.ApiUserRoleAdmin),
						string(api.ApiUserRoleReadOnly),
						string(api.ApiUserRoleUser),
					),
				},
			},
			"username": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
		},
		Description:         "ApiUsers are used to programatically interact with your Tenant",
		MarkdownDescription: "ApiUsers are used to programatically interact with your Tenant",
	}, nil
}

func (r *ApiUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("username"), req, resp)
}

func (r *ApiUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_user"
}

func (r *ApiUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiUserModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadApiUser(ctx, r.data.Client, r.data.Tenant, state.Username.Value); err != nil {
		resp.Diagnostics.AddError("Error reading ApiUser", err.Error())
		return
	} else if echoResp.GetApiUser == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		state.AppsyncEndpoint = types.String{Value: echoResp.GetApiUser.AppsyncEndpoint}
		state.Credentials = types.Object{
			Attrs: common.CognitoCredentialsAttrValues(
				echoResp.GetApiUser.Credentials.ClientId,
				echoResp.GetApiUser.Credentials.Password,
				echoResp.GetApiUser.Credentials.UserPoolId,
				echoResp.GetApiUser.Credentials.Username,
			),
			AttrTypes: common.CognitoCredentialsAttrTypes(),
		}
		if echoResp.GetApiUser.Description != nil {
			state.Description = types.String{Value: *echoResp.GetApiUser.Description}
		} else {
			state.Description = types.String{Null: true}
		}
		state.Role = types.String{Value: string(echoResp.GetApiUser.Role)}
		state.Username = types.String{Value: echoResp.GetApiUser.Username}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApiUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
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
		description = &plan.Description.Value
	}

	echoResp, err := api.UpdateApiUser(
		ctx,
		r.data.Client,
		r.data.Tenant,
		state.Username.Value,
		description,
		(*api.ApiUserRole)(&plan.Role.Value),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating ApiUser", err.Error())
		return
	}

	plan.AppsyncEndpoint = types.String{Value: echoResp.GetApiUser.Update.AppsyncEndpoint}
	plan.Credentials = types.Object{
		Attrs: common.CognitoCredentialsAttrValues(
			echoResp.GetApiUser.Update.Credentials.ClientId,
			echoResp.GetApiUser.Update.Credentials.Password,
			echoResp.GetApiUser.Update.Credentials.UserPoolId,
			echoResp.GetApiUser.Update.Credentials.Username,
		),
		AttrTypes: common.CognitoCredentialsAttrTypes(),
	}
	if echoResp.GetApiUser.Update.Description != nil {
		plan.Description = types.String{Value: *echoResp.GetApiUser.Update.Description}
	} else {
		plan.Description = types.String{Null: true}
	}
	plan.Role = types.String{Value: string(echoResp.GetApiUser.Update.Role)}
	plan.Username = types.String{Value: echoResp.GetApiUser.Update.Username}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
