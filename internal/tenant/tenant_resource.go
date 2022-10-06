package tenant

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.ResourceWithImportState = &TenantResource{}

// TenantResource defines the resource implementation.
type TenantResource struct {
	data *common.ProviderData
}

func (r *TenantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *TenantResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:          tenantResourceSchema(),
		Description:         "Manages the current Tenant",
		MarkdownDescription: "Manages the current Tenant",
	}, nil
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

	if err := r.createOrUpdate(ctx, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Unexpected error creating Tenant",
			fmt.Sprintf("This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error()),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := readTenantData(ctx, r.data.Client, r.data.Tenant, &state); err != nil {
		resp.Diagnostics.AddError(
			"Unexpected error reading Tenant",
			fmt.Sprintf("This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error()),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tenantModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.createOrUpdate(ctx, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Unexpected error updating Tenant",
			fmt.Sprintf("This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error()),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.UpdateTenant(ctx, r.data.Client, r.data.Tenant, nil, nil); err != nil {
		resp.Diagnostics.AddError("Error deleting Tenant", err.Error())
	}
}

func (r *TenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &tenantModel{})...)
}

func (r *TenantResource) createOrUpdate(ctx context.Context, data *tenantModel) error {
	var (
		config      *string
		description *string
	)

	if !data.Description.IsNull() {
		description = &data.Description.Value
	}
	if !data.Config.IsNull() {
		config = &data.Config.Value
	}
	echoResp, err := api.UpdateTenant(ctx, r.data.Client, r.data.Tenant, config, description)
	if err != nil {
		return err
	}

	if echoResp.GetTenant.Update.Description != nil {
		data.Description = types.String{Value: *echoResp.GetTenant.Update.Description}
	} else {
		data.Description = types.String{Null: true}
	}
	if echoResp.GetTenant.Update.Config != nil {
		data.Config = common.ConfigValue{Value: *echoResp.GetTenant.Update.Config}
	} else {
		data.Config = common.ConfigValue{Null: true}
	}

	return nil
}
