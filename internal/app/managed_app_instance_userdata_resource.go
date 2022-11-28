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
	_ resource.Resource = &ManagedAppInstanceUserdataResource{}
)

// ManagedAppResource defines the resource implementation.
type ManagedAppInstanceUserdataResource struct {
	data *common.ProviderData
}

type managedAppInstanceUserdataModel struct {
	App      types.String `tfsdk:"app"`
	Name     types.String `tfsdk:"name"`
	Userdata types.String `tfsdk:"userdata"`
}

func (r *ManagedAppInstanceUserdataResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ManagedAppInstanceUserdataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan managedAppInstanceUserdataModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadManagedAppUserdata(
		ctx,
		r.data.Client,
		plan.App.ValueString(),
		r.data.Tenant,
	); err != nil {
		resp.Diagnostics.AddError("Error creating ManagedAppInstanceUserdata", err.Error())
		return
	} else {
		if echoResp.GetApp != nil {
			switch app := (*echoResp.GetApp).(type) {
			case *api.ReadManagedAppUserdataGetAppManagedApp:
				plan.Userdata = types.StringValue(app.Userdata)
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

func (r *ManagedAppInstanceUserdataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *ManagedAppInstanceUserdataResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := managedAppInstanceSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"userdata": {
				Computed:            true,
				MarkdownDescription: "Cloud-init userdata specifically targeted for Amazon Linux 2.",
				Type:                types.StringType,
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		MarkdownDescription: "ManagedAppInstanceUserdata may be used to create ManagedApp compute resources based on Amazon Linux 2.",
	}, nil
}

func (r *ManagedAppInstanceUserdataResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_app_instance_userdata"
}

func (r *ManagedAppInstanceUserdataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state managedAppInstanceUserdataModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ManagedAppInstanceUserdataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state managedAppInstanceUserdataModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
