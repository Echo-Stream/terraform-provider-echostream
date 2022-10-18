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
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithImportState = &ExternalAppResource{}
	_ resource.ResourceWithModifyPlan  = &ExternalAppResource{}
)

// ExternalAppResource defines the resource implementation.
type ExternalAppResource struct {
	data *common.ProviderData
}

type externalAppModel struct {
	AppsyncEndpoint      types.String  `tfsdk:"appsync_endpoint"`
	AuditRecordsEndpoint types.String  `tfsdk:"audit_records_endpoint"`
	Config               common.Config `tfsdk:"config"`
	Credentials          types.Object  `tfsdk:"credentials"`
	Description          types.String  `tfsdk:"description"`
	Name                 types.String  `tfsdk:"name"`
	TableAccess          types.Bool    `tfsdk:"table_access"`
}

func (r *ExternalAppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ExternalAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan externalAppModel

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

	if echoResp, err := api.CreateExternalApp(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		config,
		description,
		tableAccess,
	); err != nil {
		resp.Diagnostics.AddError("Error creating ExternalApp", err.Error())
		return
	} else {
		plan.AppsyncEndpoint = types.String{Value: echoResp.CreateExternalApp.AppsyncEndpoint}
		plan.AuditRecordsEndpoint = types.String{Value: echoResp.CreateExternalApp.AuditRecordsEndpoint}
		if echoResp.CreateExternalApp.Config != nil {
			plan.Config = common.Config{Value: *echoResp.CreateExternalApp.Config}
		} else {
			plan.Config = common.Config{Null: true}
		}
		plan.Credentials = types.Object{
			Attrs: common.CognitoCredentialsAttrValues(
				echoResp.CreateExternalApp.Credentials.ClientId,
				echoResp.CreateExternalApp.Credentials.Password,
				echoResp.CreateExternalApp.Credentials.UserPoolId,
				echoResp.CreateExternalApp.Credentials.Username,
			),
			AttrTypes: common.CognitoCredentialsAttrTypes(),
		}
		if echoResp.CreateExternalApp.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateExternalApp.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		plan.Name = types.String{Value: echoResp.CreateExternalApp.Name}
		plan.TableAccess = types.Bool{Value: echoResp.CreateExternalApp.TableAccess}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ExternalAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state externalAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteApp(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ExternalApp", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ExternalAppResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := remoteAppSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"appsync_endpoint": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		Description:         "ExternalApps provide a way to process messages in their Nodes using any compute resource",
		MarkdownDescription: "ExternalApps provide a way to process messages in their Nodes using any compute resource",
	}, nil
}

func (r *ExternalAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ExternalAppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_app"
}

func (r *ExternalAppResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state externalAppModel

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
			"Destroying connected ExternalApp",
			"This will terminate the connection with your remote compute resources permanently!!",
		)
	} else {
		var plan externalAppModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.Name.Equal(state.Name) {
			resp.Diagnostics.AddWarning(
				"Replacing connected ExternalApp",
				"This will terminate the connection with your remote compute resources permanently!!",
			)
		}
	}
}

func (r *ExternalAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state externalAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadApp(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ExternalApp", err.Error())
		return
	} else if echoResp.GetApp != nil {
		switch app := (*echoResp.GetApp).(type) {
		case *api.ReadAppGetAppExternalApp:
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
			state.Name = types.String{Value: app.Name}
			state.TableAccess = types.Bool{Value: app.TableAccess}
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

func (r *ExternalAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan externalAppModel

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
		resp.Diagnostics.AddError("Error updating ExternalApp", err.Error())
		return
	}

	switch app := (*echoResp.GetApp).(type) {
	case *api.UpdateRemotetAppGetAppExternalApp:
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
