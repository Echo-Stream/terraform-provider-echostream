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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithImportState = &CrossAccountAppResource{}
	_ resource.ResourceWithModifyPlan  = &CrossAccountAppResource{}
	_ resource.ResourceWithSchema      = &CrossAccountAppResource{}
)

// CrossAccountAppResource defines the resource implementation.
type CrossAccountAppResource struct {
	data *common.ProviderData
}

type crossAccountAppModel struct {
	Account              types.String  `tfsdk:"account"`
	AppsyncEndpoint      types.String  `tfsdk:"appsync_endpoint"`
	AuditRecordsEndpoint types.String  `tfsdk:"audit_records_endpoint"`
	Config               common.Config `tfsdk:"config"`
	Credentials          types.Object  `tfsdk:"credentials"`
	Description          types.String  `tfsdk:"description"`
	IamPolicy            types.String  `tfsdk:"iam_policy"`
	Name                 types.String  `tfsdk:"name"`
	TableAccess          types.Bool    `tfsdk:"table_access"`
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
	var plan crossAccountAppModel

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

	if echoResp, err := api.CreateCrossAccountApp(
		ctx,
		r.data.Client,
		plan.Account.ValueString(),
		plan.Name.ValueString(),
		r.data.Tenant,
		config,
		description,
		tableAccess,
	); err != nil {
		resp.Diagnostics.AddError("Error creating CrossAccountApp", err.Error())
		return
	} else {
		plan.Account = types.StringValue(echoResp.CreateCrossAccountApp.Account)
		plan.AppsyncEndpoint = types.StringValue(echoResp.CreateCrossAccountApp.AppsyncEndpoint)
		plan.AuditRecordsEndpoint = types.StringValue(echoResp.CreateCrossAccountApp.AuditRecordsEndpoint)
		if echoResp.CreateCrossAccountApp.Config != nil {
			plan.Config = common.ConfigValue(*echoResp.CreateCrossAccountApp.Config)
		} else {
			plan.Config = common.ConfigNull()
		}
		var diags diag.Diagnostics
		plan.Credentials, diags = types.ObjectValue(
			common.CognitoCredentialsAttrTypes(),
			common.CognitoCredentialsAttrValues(
				echoResp.CreateCrossAccountApp.Credentials.ClientId,
				echoResp.CreateCrossAccountApp.Credentials.Password,
				echoResp.CreateCrossAccountApp.Credentials.UserPoolId,
				echoResp.CreateCrossAccountApp.Credentials.Username,
			),
		)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
		if echoResp.CreateCrossAccountApp.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateCrossAccountApp.Description)
		} else {
			plan.Description = types.StringNull()
		}
		plan.IamPolicy = types.StringValue(echoResp.CreateCrossAccountApp.IamPolicy)
		plan.Name = types.StringValue(echoResp.CreateCrossAccountApp.Name)
		plan.TableAccess = types.BoolValue(echoResp.CreateCrossAccountApp.TableAccess)
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

	if _, err := api.DeleteApp(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting CrossAccountApp", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
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
			fmt.Sprintf("This will terminate the connection with %s permanently!!", state.Account.ValueString()),
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
				fmt.Sprintf("This will terminate the connection with %s permanently!!", state.Account.ValueString()),
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

	if echoResp, err := api.ReadApp(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossAccountApp", err.Error())
		return
	} else if echoResp.GetApp != nil {
		switch app := (*echoResp.GetApp).(type) {
		case *api.ReadAppGetAppCrossAccountApp:
			state.Account = types.StringValue(app.Account)
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
			state.IamPolicy = types.StringValue(app.IamPolicy)
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

func (r *CrossAccountAppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := remoteAppAttributes()
	maps.Copy(
		attributes,
		map[string]schema.Attribute{
			"account": schema.StringAttribute{
				MarkdownDescription: "The AWS account number that will host this CrossAcountApp's compute resources.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(12, 12),
					stringvalidator.RegexMatches(
						regexp.MustCompile("^[0-9]+$"),
						"value must contain only numbers",
					),
				},
			},
			"appsync_endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The EchoStream AppSync Endpoint that this ExternalApp must use.",
			},
			"iam_policy": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The IAM policy to apply to this CrossAccountApp's compute resources (e.g. - Lambda, EC2) to grant access to its EchoStream resources.",
			},
		},
	)
	resp.Schema = schema.Schema{
		Attributes:          attributes,
		MarkdownDescription: "[CrossAccountApps](https://docs.echo.stream/docs/cross-account-app) provides a way to receive/send messages in their Nodes using cross-account IAM access.",
	}
}

func (r *CrossAccountAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan crossAccountAppModel

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
		resp.Diagnostics.AddError("Error updating CrossAccountApp", err.Error())
		return
	}

	switch app := (*echoResp.GetApp).(type) {
	case *api.UpdateRemotetAppGetAppCrossAccountApp:
		plan.Account = types.StringValue(app.Update.Account)
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
		plan.IamPolicy = types.StringValue(app.Update.IamPolicy)
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
