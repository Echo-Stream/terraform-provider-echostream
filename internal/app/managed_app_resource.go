package app

import (
	"context"
	"fmt"
	"regexp"
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

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithImportState = &ManagedAppResource{}
	_ resource.ResourceWithModifyPlan  = &ManagedAppResource{}
)

// ManagedAppResource defines the resource implementation.
type ManagedAppResource struct {
	data *common.ProviderData
}

type managedAppModel struct {
	AuditRecordsEndpoint types.String  `tfsdk:"audit_records_endpoint"`
	Config               common.Config `tfsdk:"config"`
	Credentials          types.Object  `tfsdk:"credentials"`
	Description          types.String  `tfsdk:"description"`
	Name                 types.String  `tfsdk:"name"`
	TableAccess          types.Bool    `tfsdk:"table_access"`
}

func (r *ManagedAppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ManagedAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan managedAppModel

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
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.TableAccess.IsNull() || plan.TableAccess.IsUnknown()) {
		tableAccess = &plan.TableAccess.Value
	}

	if echoResp, err := api.CreateManagedApp(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		config,
		description,
		tableAccess,
	); err != nil {
		resp.Diagnostics.AddError("Error creating ManagedApp", err.Error())
		return
	} else {
		plan.AuditRecordsEndpoint = types.String{Value: echoResp.CreateManagedApp.AuditRecordsEndpoint}
		if echoResp.CreateManagedApp.Config != nil {
			plan.Config = common.Config{Value: *echoResp.CreateManagedApp.Config}
		} else {
			plan.Config = common.Config{Null: true}
		}
		plan.Credentials = types.Object{
			Attrs: common.CognitoCredentialsAttrValues(
				echoResp.CreateManagedApp.Credentials.ClientId,
				echoResp.CreateManagedApp.Credentials.Password,
				echoResp.CreateManagedApp.Credentials.UserPoolId,
				echoResp.CreateManagedApp.Credentials.Username,
			),
			AttrTypes: common.CognitoCredentialsAttrTypes(),
		}
		if echoResp.CreateManagedApp.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateManagedApp.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		plan.Name = types.String{Value: echoResp.CreateManagedApp.Name}
		plan.TableAccess = types.Bool{Value: echoResp.CreateManagedApp.TableAccess}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ManagedAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state managedAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteApp(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ManagedApp", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ManagedAppResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := remoteAppSchema()
	name := schema["name"]
	name.Validators = []tfsdk.AttributeValidator{
		stringvalidator.LengthBetween(3, 80),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[A-Za-z0-9\-\_]*$`),
			"value must contain only lowercase/uppercase alphanumeric characters, \"-\", \"_\"",
		),
	}
	return tfsdk.Schema{
		Attributes:          schema,
		Description:         "ManagedApps provide fully managed (by EchoStream) processing resources in a remote virtual compute environment",
		MarkdownDescription: "ManagedApps provide fully managed (by EchoStream) processing resources in a remote virtual compute environment",
	}, nil
}

func (r *ManagedAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ManagedAppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_app"
}

func (r *ManagedAppResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state managedAppModel

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
			"Destroying connected ManagedApp",
			"This will terminate the connection with your remote compute resources permanently!!",
		)
	} else {
		var plan managedAppModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.Name.Equal(state.Name) {
			resp.Diagnostics.AddWarning(
				"Replacing connected ManagedApp",
				"This will terminate the connection with your remote compute resources permanently!!",
			)
		}
	}
}

func (r *ManagedAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state managedAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadApp(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ManagedApp", err.Error())
		return
	} else if echoResp.GetApp != nil {
		switch app := (*echoResp.GetApp).(type) {
		case *api.ReadAppGetAppManagedApp:
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

func (r *ManagedAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan managedAppModel

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
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.TableAccess.IsNull() || plan.TableAccess.IsUnknown()) {
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
		resp.Diagnostics.AddError("Error updating ManagedApp", err.Error())
		return
	}

	switch app := (*echoResp.GetApp).(type) {
	case *api.UpdateRemotetAppGetAppManagedApp:
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
