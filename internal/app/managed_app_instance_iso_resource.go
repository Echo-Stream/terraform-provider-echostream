package app

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource = &ManagedAppInstanceIsoResource{}
)

// ManagedAppResource defines the resource implementation.
type ManagedAppInstanceIsoResource struct {
	data *common.ProviderData
}

type managedAppInstanceIsoModel struct {
	App  types.String `tfsdk:"app"`
	Name types.String `tfsdk:"name"`
	Iso  types.String `tfsdk:"iso"`
}

func (r *ManagedAppInstanceIsoResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ManagedAppInstanceIsoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan managedAppInstanceIsoModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadManagedAppIso(
		ctx,
		r.data.Client,
		plan.App.Value,
		r.data.Tenant,
	); err != nil {
		resp.Diagnostics.AddError("Error creating ManagedAppInstanceIso", err.Error())
		return
	} else {
		if echoResp.GetApp != nil {
			switch app := (*echoResp.GetApp).(type) {
			case *api.ReadManagedAppIsoGetAppManagedApp:
				plan.Iso = types.String{Value: app.Iso}
			default:
				resp.Diagnostics.AddError(
					"Incorrect App type",
					fmt.Sprintf("'%s' is incorrect App type", plan.App.String()),
				)
			}
		} else {
			resp.Diagnostics.AddError(
				"ManagedApp not found",
				fmt.Sprintf("'%s' ManagedApp does not exist", plan.App.String()),
			)
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ManagedAppInstanceIsoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *ManagedAppInstanceIsoResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := managedAppInstanceSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"iso": {
				Computed:            true,
				MarkdownDescription: "The iso image, gzip'd and base64 encoded",
				Type:                types.StringType,
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		MarkdownDescription: "ManagedAppInstanceIso may be used to create ManagedApp compute resources in the VM architecture of your choice",
	}, nil
}

func (r *ManagedAppInstanceIsoResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_app_instance_iso"
}

func (r *ManagedAppInstanceIsoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state managedAppInstanceIsoModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ManagedAppInstanceIsoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state managedAppInstanceIsoModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
