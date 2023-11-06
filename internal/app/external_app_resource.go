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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigure   = &ExternalAppResource{}
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

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.TableAccess.IsNull() || plan.TableAccess.IsUnknown()) {
		temp := plan.TableAccess.ValueBool()
		tableAccess = &temp
	}

	if echoResp, err := api.CreateExternalApp(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		r.data.Tenant,
		config,
		description,
		tableAccess,
	); err != nil {
		resp.Diagnostics.AddError("Error creating ExternalApp", err.Error())
		return
	} else {
		plan.AppsyncEndpoint = types.StringValue(echoResp.CreateExternalApp.AppsyncEndpoint)
		plan.AuditRecordsEndpoint = types.StringValue(echoResp.CreateExternalApp.AuditRecordsEndpoint)
		if echoResp.CreateExternalApp.Config != nil {
			plan.Config = common.ConfigValue(*echoResp.CreateExternalApp.Config)
		} else {
			plan.Config = common.ConfigNull()
		}
		var diags diag.Diagnostics
		plan.Credentials, diags = types.ObjectValue(
			common.CognitoCredentialsAttrTypes(),
			common.CognitoCredentialsAttrValues(
				echoResp.CreateExternalApp.Credentials.ClientId,
				echoResp.CreateExternalApp.Credentials.Password,
				echoResp.CreateExternalApp.Credentials.UserPoolId,
				echoResp.CreateExternalApp.Credentials.Username,
			),
		)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
		if echoResp.CreateExternalApp.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateExternalApp.Description)
		} else {
			plan.Description = types.StringNull()
		}
		plan.Name = types.StringValue(echoResp.CreateExternalApp.Name)
		plan.TableAccess = types.BoolValue(echoResp.CreateExternalApp.TableAccess)
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

	if _, err := api.DeleteApp(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ExternalApp", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
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

	if echoResp, err := api.ReadApp(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ExternalApp", err.Error())
		return
	} else if echoResp.GetApp != nil {
		switch app := (*echoResp.GetApp).(type) {
		case *api.ReadAppGetAppExternalApp:
			state.AppsyncEndpoint = types.StringValue(app.AppsyncEndpoint)
			state.AuditRecordsEndpoint = types.StringValue(app.AuditRecordsEndpoint)
			state.Name = types.StringValue(app.Name)
			if app.Config != nil {
				state.Config = common.ConfigValue(*app.Config)
			} else {
				state.Config = common.ConfigNull()
			}
			var diags diag.Diagnostics
			state.Credentials, diags = types.ObjectValue(
				common.CognitoCredentialsAttrTypes(),
				common.CognitoCredentialsAttrValues(
					app.Credentials.ClientId,
					app.Credentials.Password,
					app.Credentials.UserPoolId,
					app.Credentials.Username,
				),
			)
			if diags != nil && diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
			if app.Description != nil {
				state.Description = types.StringValue(*app.Description)
			} else {
				state.Description = types.StringNull()
			}
			state.Name = types.StringValue(app.Name)
			state.TableAccess = types.BoolValue(app.TableAccess)
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

func (r *ExternalAppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := remoteAppResourceAttributes()
	maps.Copy(
		attributes,
		map[string]schema.Attribute{
			"appsync_endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The EchoStream AppSync Endpoint that this ExternalApp must use.",
			},
		},
	)
	resp.Schema = schema.Schema{
		Attributes:          attributes,
		MarkdownDescription: "[ExternalApps](https://docs.echo.stream/docs/external-app) provide a way to process messages in their Nodes using any compute resource.",
	}
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

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.TableAccess.IsNull() || plan.TableAccess.IsUnknown()) {
		temp := plan.TableAccess.ValueBool()
		tableAccess = &temp
	}

	echoResp, err := api.UpdateRemotetApp(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
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
		plan.AppsyncEndpoint = types.StringValue(app.Update.AppsyncEndpoint)
		plan.AuditRecordsEndpoint = types.StringValue(app.Update.AuditRecordsEndpoint)
		plan.Name = types.StringValue(app.Update.Name)
		if app.Update.Config != nil {
			plan.Config = common.ConfigValue(*app.Update.Config)
		} else {
			plan.Config = common.ConfigNull()
		}
		var diags diag.Diagnostics
		plan.Credentials, diags = types.ObjectValue(
			common.CognitoCredentialsAttrTypes(),
			common.CognitoCredentialsAttrValues(
				app.Update.Credentials.ClientId,
				app.Update.Credentials.Password,
				app.Update.Credentials.UserPoolId,
				app.Update.Credentials.Username,
			),
		)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
		if app.Update.Description != nil {
			plan.Description = types.StringValue(*app.Update.Description)
		} else {
			plan.Description = types.StringNull()
		}
		plan.Name = types.StringValue(app.Update.Name)
		plan.TableAccess = types.BoolValue(app.Update.TableAccess)
	default:
		resp.Diagnostics.AddError(
			"Incorrect App type",
			fmt.Sprintf("'%s' is incorrect App type", plan.Name.String()),
		)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
