package tenant

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigure   = &TenantResource{}
	_ resource.ResourceWithImportState = &TenantResource{}
)

// TenantResource defines the resource implementation.
type TenantResource struct {
	data *common.ProviderData
}

func (r *TenantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.createOrUpdate(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteTenant(ctx, r.data.Client, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting Tenant", err.Error())
		return
	}
}

func (r *TenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), r.data.Tenant)...)
}

func (r *TenantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *TenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(readTenantData(ctx, r.data.Client, r.data.Tenant, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TenantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"active": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "The current Tenant's active state.",
			},
			"audit": schema.BoolAttribute{
				MarkdownDescription: "The current Tenant's audit state.",
				Optional:            true,
			},
			"aws_credentials": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"access_key_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The AWS Acces Key Id for the session.",
					},
					"expiration": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The date/time that the sesssion expires, in [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) format.",
					},
					"secret_access_key": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The AWS Secret Access Key for the session.",
						Sensitive:           true,
					},
					"session_token": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The AWS Session Token for the session.",
					},
				},
				Computed:            true,
				MarkdownDescription: "The AWS Session Credentials that allow the current ApiUser (configured in the provider) to access the Tenant's resources.",
			},
			"aws_credentials_duration": schema.Int64Attribute{
				MarkdownDescription: "The duration to request for `aws_credentials`. Must be set to obtain `aws_credentials`.",
				Optional:            true,
			},
			"config": schema.StringAttribute{
				CustomType:          common.ConfigType{},
				MarkdownDescription: "The config for the Tenant. All nodes in the Tenant will be allowed to access this. Must be a JSON object.",
				Optional:            true,
				Sensitive:           true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name.",
			},
			"region": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current Tenant's AWS region name (e.g.  - `us-east-1`).",
			},
			"table": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current Tenant's DynamoDB [table](https://docs.echo.stream/docs/table) name.",
			},
		},
		MarkdownDescription: "Manages the current [Tenant](https://docs.echo.stream/docs/tenants) (configured in the provider)",
	}
}

func (r *TenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tenantModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.createOrUpdate(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TenantResource) createOrUpdate(ctx context.Context, data *tenantModel) diag.Diagnostics {
	var (
		audit       *bool
		config      *string
		description *string
		diags       diag.Diagnostics
	)

	if !(data.Audit.IsNull() || data.Audit.IsUnknown()) {
		temp := data.Audit.ValueBool()
		audit = &temp
	}
	if !(data.Description.IsNull() || data.Description.IsUnknown()) {
		temp := data.Description.ValueString()
		description = &temp
	}
	if !(data.Config.IsNull() || data.Config.IsUnknown()) {
		temp := data.Config.ValueConfig()
		config = &temp
	}

	if echoResp, err := api.UpdateTenant(ctx, r.data.Client, r.data.Tenant, audit, config, description); err != nil {
		diags.AddError(
			"Unexpected error creating or updating Tenant",
			fmt.Sprintf("This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error()),
		)
	} else if echoResp == nil {
		diags.AddError(
			"Unexpected error creating or updating Tenant",
			fmt.Sprintf("Unable to find Tenant '%s'", r.data.Tenant),
		)
	} else {
		data.Active = types.BoolValue(echoResp.GetTenant.Update.Active)
		if echoResp.GetTenant.Update.Config != nil {
			data.Config = common.ConfigValue(*echoResp.GetTenant.Update.Config)
		} else {
			data.Config = common.ConfigNull()
		}
		if echoResp.GetTenant.Update.Description != nil {
			data.Description = types.StringValue(*echoResp.GetTenant.Update.Description)
		} else {
			data.Description = types.StringNull()
		}
		data.Id = types.StringValue(echoResp.GetTenant.Update.Name)
		data.Name = types.StringValue(echoResp.GetTenant.Update.Name)
		data.Region = types.StringValue(echoResp.GetTenant.Update.Region)
		data.Table = types.StringValue(echoResp.GetTenant.Update.Table)
		diags.Append(readTenantAwsCredentials(ctx, r.data.Client, r.data.Tenant, data)...)
	}

	return diags
}
