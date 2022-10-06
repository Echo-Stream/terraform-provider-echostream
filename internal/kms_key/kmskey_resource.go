package kmskey

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithImportState = &KmsKeyResource{}

type KmsKeyResource struct {
	data *common.ProviderData
}

type kmsKeyModel struct {
	Arn         types.String `tfsdk:"arn"`
	Description types.String `tfsdk:"description"`
	InUse       types.Bool   `tfsdk:"in_use"`
	Name        types.String `tfsdk:"name"`
}

func (r *KmsKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KmsKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan kmsKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string

	if !plan.Description.IsNull() {
		description = &plan.Description.Value
	}

	echoResp, err := api.CreateKmsKey(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		description,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating KmsKey", err.Error())
		return
	}

	plan.Arn = types.String{Value: echoResp.CreateKmsKey.Arn}
	if echoResp.CreateKmsKey.Description != nil {
		plan.Description = types.String{Value: *echoResp.CreateKmsKey.Description}
	} else {
		plan.Description = types.String{Null: true}
	}
	plan.InUse = types.Bool{Value: echoResp.CreateKmsKey.InUse}
	plan.Name = types.String{Value: echoResp.CreateKmsKey.Name}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *KmsKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state kmsKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteKmsKey(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting KmsKey", err.Error())
		return
	}
}

func (r *KmsKeyResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"arn": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
			"description": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
			},
			"in_use": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.BoolType,
			},
			"name": {
				Description:         "",
				MarkdownDescription: "",
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Required:            true,
				Type:                types.StringType,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.LengthBetween(3, 80),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9\-\_]*$`),
						"value must contain only lowercase/uppercase alphanumeric characters, \"-\" or \"_\"",
					),
				},
			},
		},
		Description:         "KmsKeys are used to encrypt message on Edges",
		MarkdownDescription: "KmsKeys are used to encrypt message on Edges",
	}, nil
}

func (r *KmsKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *KmsKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms_key"
}

func (r *KmsKeyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state kmsKeyModel

	// If the entire state is null, resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If the KmsKeyis not in use it can be destroyed at will.
	if !state.InUse.Value {
		return
	}

	// If the entire plan is null, the resource is planned for destruction.
	prevent_destroy := req.Plan.Raw.IsNull()
	if !prevent_destroy {
		var plan kmsKeyModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		prevent_destroy = !plan.Name.Equal(state.Name)
	}

	if prevent_destroy {
		resp.Diagnostics.AddError("Cannot destroy KmsKey", fmt.Sprintf("KmsKey %s is in use and may not be destroyed", state.Name.Value))
	}
}

func (r *KmsKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state kmsKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	echoResp, err := api.ReadKmsKey(ctx, r.data.Client, state.Name.Value, r.data.Tenant)
	if err != nil {
		resp.Diagnostics.AddError("Error reading KmsKey", err.Error())
		return
	}

	state.Arn = types.String{Value: echoResp.GetKmsKey.Name}
	if echoResp.GetKmsKey.Description != nil {
		state.Description = types.String{Value: *echoResp.GetKmsKey.Description}
	} else {
		state.Description = types.String{Null: true}
	}
	state.InUse = types.Bool{Value: echoResp.GetKmsKey.InUse}
	state.Name = types.String{Value: echoResp.GetKmsKey.Name}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KmsKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *kmsKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string
	if !plan.Description.IsNull() {
		description = &plan.Description.Value
	}

	echoResp, err := api.UpdateKmsKey(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		description,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating KmsKey", err.Error())
		return
	}

	plan.Arn = types.String{Value: echoResp.GetKmsKey.Update.Arn}
	if echoResp.GetKmsKey.Update.Description != nil {
		plan.Description = types.String{Value: *echoResp.GetKmsKey.Update.Description}
	} else {
		plan.Description = types.String{Null: true}
	}
	plan.InUse = types.Bool{Value: echoResp.GetKmsKey.Update.InUse}
	plan.Name = types.String{Value: echoResp.GetKmsKey.Update.Name}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
