package app

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
	_ resource.ResourceWithImportState = &CrossAccountAppResource{}
	_ resource.ResourceWithModifyPlan  = &CrossAccountAppResource{}
)

// CrossAccountAppResource defines the resource implementation.
type CrossAccountAppResource struct {
	data *common.ProviderData
}

func (r *CrossAccountAppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CrossAccountAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *crossAccountAppModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config      *string
		description *string
		tableAccess *bool
	)

	if !plan.Config.IsNull() {
		config = &plan.Config.Value
	}
	if !plan.Description.IsNull() {
		description = &plan.Description.Value
	}
	if !plan.TableAccess.IsNull() {
		tableAccess = &plan.TableAccess.Value
	}

	if echoResp, err := api.CreateCrossAccountApp(
		ctx,
		r.data.Client,
		plan.Account.Value,
		plan.Name.Value,
		r.data.Tenant,
		config,
		description,
		tableAccess,
	); err != nil {
		resp.Diagnostics.AddError("Error creating CrossAccountApp", err.Error())
		return
	} else {
		plan.Account = types.String{Value: echoResp.CreateCrossAccountApp.Account}
		plan.AppsyncEndpoint = types.String{Value: echoResp.CreateCrossAccountApp.AppsyncEndpoint}
		plan.AuditRecordsEndpoint = types.String{Value: echoResp.CreateCrossAccountApp.AuditRecordsEndpoint}
		if echoResp.CreateCrossAccountApp.Config != nil {
			plan.Config = common.Config{Value: *echoResp.CreateCrossAccountApp.Config}
		} else {
			plan.Config = common.Config{Null: true}
		}
		plan.Credentials = types.Object{
			Attrs: common.CognitoCredentialsAttrValues(
				echoResp.CreateCrossAccountApp.Credentials.ClientId,
				echoResp.CreateCrossAccountApp.Credentials.Password,
				echoResp.CreateCrossAccountApp.Credentials.UserPoolId,
				echoResp.CreateCrossAccountApp.Credentials.Username,
			),
			AttrTypes: common.CognitoCredentialsAttrTypes(),
		}
		if echoResp.CreateCrossAccountApp.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateCrossAccountApp.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		plan.IamPolicy = types.String{Value: echoResp.CreateCrossAccountApp.IamPolicy}
		plan.Name = types.String{Value: echoResp.CreateCrossAccountApp.Name}
		plan.TableAccess = types.Bool{Value: echoResp.CreateCrossAccountApp.TableAccess}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CrossAccountAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state crossAccountAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteApp(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting CrossAccountApp", err.Error())
		return
	}
}

func (r *CrossAccountAppResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:          crossAccountAppSchema(),
		Description:         "CrossAccountApps provide a way to receive/send messages in their Nodes using cross-account IAM access",
		MarkdownDescription: "CrossAccountApps provide a way to receive/send messages in their Nodes using cross-account IAM access",
	}, nil
}

func (r *CrossAccountAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *CrossAccountAppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cross_account_app"
}

func (r *CrossAccountAppResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state crossAccountAppModel

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
			"Destroying connected CrossAccountApp",
			fmt.Sprintf("This will terminate the connection with %s permanently!!", state.Account.Value),
		)
	} else {
		var plan crossAccountAppModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		if !(plan.Name.Equal(state.Name) && plan.Account.Equal(state.Account)) {
			resp.Diagnostics.AddWarning(
				"Replacing connected CrossAccountApp",
				fmt.Sprintf("This will terminate the connection with %s permanently!!", state.Account.Value),
			)
		}
	}
}

func (r *CrossAccountAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state crossAccountAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadApp(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossAccountApp", err.Error())
		return
	} else if echoResp.GetApp != nil {
		switch app := (*echoResp.GetApp).(type) {
		case *api.ReadAppGetAppCrossAccountApp:
			state.Account = types.String{Value: app.Account}
			state.AppsyncEndpoint = types.String{Value: app.AppsyncEndpoint}
			state.AuditRecordsEndpoint = types.String{Value: app.AuditRecordsEndpoint}
			state.Name = types.String{Value: app.Name}
			if app.Config != nil {
				state.Config = common.Config{Value: *app.Config}
			} else {
				state.Config = common.Config{Null: true}
			}
			state.Credentials = types.Object{
				Attrs: common.CognitoCredentialsAttrValues(
					app.Credentials.ClientId,
					app.Credentials.Password,
					app.Credentials.UserPoolId,
					app.Credentials.Username,
				),
				AttrTypes: common.CognitoCredentialsAttrTypes(),
			}
			if app.Description != nil {
				state.Description = types.String{Value: *app.Description}
			} else {
				state.Description = types.String{Null: true}
			}
			state.IamPolicy = types.String{Value: app.IamPolicy}
			state.Name = types.String{Value: app.Name}
			state.TableAccess = types.Bool{Value: app.TableAccess}
		default:
			resp.Diagnostics.AddError(
				"Incorrect App type",
				fmt.Sprintf("'%s' is incorrect App type", state.Name.String()),
			)
		}
	} else {
		resp.Diagnostics.AddError(
			"CrossAccountApp not found",
			fmt.Sprintf("'%s' CrossAccountApp does not exist", state.Name.String()),
		)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CrossAccountAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *crossAccountAppModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config      *string
		description *string
		tableAccess *bool
	)

	if !plan.Config.IsNull() {
		config = &plan.Config.Value
	}
	if !plan.Description.IsNull() {
		description = &plan.Description.Value
	}
	if !plan.TableAccess.IsNull() {
		tableAccess = &plan.TableAccess.Value
	}

	echoResp, err := api.UpdateRemotetApp(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		config,
		description,
		tableAccess,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating CrossAccountApp", err.Error())
		return
	}

	switch app := (*echoResp.GetApp).(type) {
	case *api.UpdateRemotetAppGetAppCrossAccountApp:
		plan.Account = types.String{Value: app.Update.Account}
		plan.AppsyncEndpoint = types.String{Value: app.Update.AppsyncEndpoint}
		plan.AuditRecordsEndpoint = types.String{Value: app.Update.AuditRecordsEndpoint}
		plan.Name = types.String{Value: app.Update.Name}
		if app.Update.Config != nil {
			plan.Config = common.Config{Value: *app.Update.Config}
		} else {
			plan.Config = common.Config{Null: true}
		}
		plan.Credentials = types.Object{
			Attrs: common.CognitoCredentialsAttrValues(
				app.Update.Credentials.ClientId,
				app.Update.Credentials.Password,
				app.Update.Credentials.UserPoolId,
				app.Update.Credentials.Username,
			),
			AttrTypes: common.CognitoCredentialsAttrTypes(),
		}
		if app.Update.Description != nil {
			plan.Description = types.String{Value: *app.Update.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		plan.IamPolicy = types.String{Value: app.Update.IamPolicy}
		plan.Name = types.String{Value: app.Update.Name}
		plan.TableAccess = types.Bool{Value: app.Update.TableAccess}
	default:
		resp.Diagnostics.AddError(
			"Incorrect App type",
			fmt.Sprintf("'%s' is incorrect App type", plan.Name.String()),
		)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
