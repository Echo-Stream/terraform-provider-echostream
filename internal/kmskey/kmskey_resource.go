package kmskey

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &KmsKeyResource{}
	_ resource.ResourceWithImportState = &KmsKeyResource{}
	_ resource.ResourceWithModifyPlan  = &KmsKeyResource{}
)

type KmsKeyResource struct {
	data *common.ProviderData
}

type kmsKeyModel struct {
	Arn         types.String `tfsdk:"arn"`
	Description types.String `tfsdk:"description"`
	Id          types.String `tfsdk:"id"`
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

	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}

	echoResp, err := api.CreateKmsKey(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		r.data.Tenant,
		description,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating KmsKey", err.Error())
		return
	}

	plan.Arn = types.StringValue(echoResp.CreateKmsKey.Arn)
	if echoResp.CreateKmsKey.Description != nil {
		plan.Description = types.StringValue(*echoResp.CreateKmsKey.Description)
	} else {
		plan.Description = types.StringNull()
	}
	plan.Id = types.StringValue(echoResp.CreateKmsKey.Name)
	plan.InUse = types.BoolValue(echoResp.CreateKmsKey.InUse)
	plan.Name = types.StringValue(echoResp.CreateKmsKey.Name)

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

	if _, err := api.DeleteKmsKey(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting KmsKey", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
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
	if !state.InUse.ValueBool() {
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
		resp.Diagnostics.AddError("Cannot destroy KmsKey", fmt.Sprintf("KmsKey %s is in use and may not be destroyed", state.Name.ValueString()))
	}
}

func (r *KmsKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state kmsKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadKmsKey(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading KmsKey", err.Error())
		return
	} else if echoResp.GetKmsKey == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		state.Arn = types.StringValue(echoResp.GetKmsKey.Name)
		if echoResp.GetKmsKey.Description != nil {
			state.Description = types.StringValue(*echoResp.GetKmsKey.Description)
		} else {
			state.Description = types.StringNull()
		}
		state.Id = types.StringValue(echoResp.GetKmsKey.Name)
		state.InUse = types.BoolValue(echoResp.GetKmsKey.InUse)
		state.Name = types.StringValue(echoResp.GetKmsKey.Name)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KmsKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The AWS ARN for the underlying KMS Key.",
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"in_use": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "True if this KmsKey is in use by Edges.",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the KmsKey. Must be unique within the Tenant.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 80),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9\-\_]*$`),
						"value must contain only lowercase/uppercase alphanumeric characters, \"-\" or \"_\"",
					),
				},
			},
		},
		MarkdownDescription: "KmsKeys are used to encrypt message on Edges. This enables limiting certain Apps and Nodes, " +
			"especially External and Managed Nodes that are outside of your control (e.g. - shared with a partner), to specific encryption permissions.",
	}
}

func (r *KmsKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan kmsKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}

	echoResp, err := api.UpdateKmsKey(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		r.data.Tenant,
		description,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating KmsKey", err.Error())
		return
	}

	plan.Arn = types.StringValue(echoResp.GetKmsKey.Update.Arn)
	if echoResp.GetKmsKey.Update.Description != nil {
		plan.Description = types.StringValue(*echoResp.GetKmsKey.Update.Description)
	} else {
		plan.Description = types.StringNull()
	}
	plan.Id = types.StringValue(echoResp.GetKmsKey.Update.Name)
	plan.InUse = types.BoolValue(echoResp.GetKmsKey.Update.InUse)
	plan.Name = types.StringValue(echoResp.GetKmsKey.Update.Name)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
