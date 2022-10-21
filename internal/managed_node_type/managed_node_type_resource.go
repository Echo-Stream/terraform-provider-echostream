package managed_node_type

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
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithImportState = &ManagedNodeTypeResource{}
	_ resource.ResourceWithModifyPlan  = &ManagedNodeTypeResource{}
)

// ManagedNodeTypeResource defines the resource implementation.
type ManagedNodeTypeResource struct {
	data *common.ProviderData
}

func (r *ManagedNodeTypeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ManagedNodeTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan managedNodeTypeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		configTemplate     *string
		mountRequirements  []api.MountRequirementInput
		portRequirements   []api.PortRequirementInput
		readme             *string
		receiveMessageType *string
		sendMessageType    *string
	)

	if !(plan.ConfigTemplate.IsNull() || plan.ConfigTemplate.IsUnknown()) {
		configTemplate = &plan.ConfigTemplate.Value
	}
	if !(plan.MountRequirements.IsNull() || plan.MountRequirements.IsUnknown()) {
		mr := []mountRequirementsModel{}
		resp.Diagnostics.Append(plan.MountRequirements.ElementsAs(ctx, &mr, false)...)
		if !resp.Diagnostics.HasError() {
			mountRequirements = make([]api.MountRequirementInput, len(mr))
			for i, t_mri := range mr {
				mri := api.MountRequirementInput{
					Description: t_mri.Description.Value,
					Target:      t_mri.Target.Value,
				}
				if !(t_mri.Source.IsNull() || t_mri.Source.IsUnknown()) {
					mri.Source = &t_mri.Source.Value
				}
				mountRequirements[i] = mri
			}
		}
	}
	if !(plan.PortRequirements.IsNull() || plan.PortRequirements.IsUnknown()) {
		pr := []portRequirementsModel{}
		resp.Diagnostics.Append(plan.PortRequirements.ElementsAs(ctx, &pr, false)...)
		if !resp.Diagnostics.HasError() {
			portRequirements = make([]api.PortRequirementInput, len(pr))
			for i, t_pri := range pr {
				pri := api.PortRequirementInput{
					ContainerPort: int(t_pri.ContainerPort.Value),
					Description:   t_pri.Description.Value,
					Protocol:      api.Protocol(t_pri.Protocol.Value),
				}
				portRequirements[i] = pri
			}
		}
	}
	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		readme = &plan.Readme.Value
	}
	if !(plan.ReceiveMessageType.IsNull() || plan.ReceiveMessageType.IsUnknown()) {
		receiveMessageType = &plan.ReceiveMessageType.Value
	}
	if !(plan.SendMessageType.IsNull() || plan.SendMessageType.IsUnknown()) {
		sendMessageType = &plan.SendMessageType.Value
	}

	if resp.Diagnostics.HasError() {
		return
	}

	echoResp, err := api.CreateManagedNodeType(
		ctx,
		r.data.Client,
		plan.Description.Value,
		plan.ImageUri.Value,
		plan.Name.Value,
		r.data.Tenant,
		configTemplate,
		mountRequirements,
		portRequirements,
		readme,
		receiveMessageType,
		sendMessageType,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating ManagedNodeType", err.Error())
		return
	}

	if echoResp.CreateManagedNodeType.ConfigTemplate != nil {
		plan.ConfigTemplate = common.Config{Value: *echoResp.CreateManagedNodeType.ConfigTemplate}
	} else {
		plan.ConfigTemplate = common.Config{Null: true}
	}
	plan.Description = types.String{Value: echoResp.CreateManagedNodeType.Description}
	plan.ImageUri = types.String{Value: echoResp.CreateManagedNodeType.ImageUri}
	plan.InUse = types.Bool{Value: echoResp.CreateManagedNodeType.InUse}
	plan.MountRequirements = types.Set{ElemType: types.ObjectType{AttrTypes: mountRequirementsAttrTypes()}}
	if len(echoResp.CreateManagedNodeType.MountRequirements) > 0 {
		for _, mr := range echoResp.CreateManagedNodeType.MountRequirements {
			plan.MountRequirements.Elems = append(
				plan.MountRequirements.Elems,
				types.Object{
					Attrs:     mountRequirementsAttrValues(mr.Description, mr.Source, mr.Target),
					AttrTypes: mountRequirementsAttrTypes(),
				},
			)
		}
	} else {
		plan.MountRequirements.Null = true
	}
	plan.Name = types.String{Value: echoResp.CreateManagedNodeType.Name}
	plan.PortRequirements = types.Set{ElemType: types.ObjectType{AttrTypes: portRequirementAttrTypes()}}
	if len(echoResp.CreateManagedNodeType.PortRequirements) > 0 {
		for _, pr := range echoResp.CreateManagedNodeType.PortRequirements {
			plan.PortRequirements.Elems = append(
				plan.PortRequirements.Elems,
				types.Object{
					Attrs:     portRequirementAttrValues(pr.ContainerPort, pr.Description, pr.Protocol),
					AttrTypes: portRequirementAttrTypes(),
				},
			)
		}
	} else {
		plan.PortRequirements.Null = true
	}
	if echoResp.CreateManagedNodeType.Readme != nil {
		plan.Readme = types.String{Value: *echoResp.CreateManagedNodeType.Readme}
	} else {
		plan.Readme = types.String{Null: true}
	}
	if echoResp.CreateManagedNodeType.ReceiveMessageType != nil {
		plan.ReceiveMessageType = types.String{Value: echoResp.CreateManagedNodeType.ReceiveMessageType.Name}
	} else {
		plan.ReceiveMessageType = types.String{Null: true}
	}
	if echoResp.CreateManagedNodeType.SendMessageType != nil {
		plan.SendMessageType = types.String{Value: echoResp.CreateManagedNodeType.SendMessageType.Name}
	} else {
		plan.SendMessageType = types.String{Null: true}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ManagedNodeTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state managedNodeTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteManagedNodeType(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ManagedNodeType", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ManagedNodeTypeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:          resourceManagedNodeTypeSchema(),
		Description:         "ManagedNodeTypes are used to define ManagedNodes",
		MarkdownDescription: "ManagedNodeTypes are used to define ManagedNodes",
	}, nil
}

func (r *ManagedNodeTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ManagedNodeTypeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_node_type"
}

func (r *ManagedNodeTypeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state managedNodeTypeModel

	// If the entire state is null, resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If the ManagedNodeType is not in use it can be destroyed at will.
	if !state.InUse.Value {
		return
	}

	// If the entire plan is null, the resource is planned for destruction.
	prevent_destroy := req.Plan.Raw.IsNull()
	if !prevent_destroy {
		var plan managedNodeTypeModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		prevent_destroy = !plan.Name.Equal(state.Name)
	}

	if prevent_destroy {
		resp.Diagnostics.AddError("Cannot destroy ManagedNodeType", fmt.Sprintf("ManagedNodeType %s is in use and may not be destroyed", state.Name.Value))
	}
}

func (r *ManagedNodeTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state managedNodeTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, system, err := readManagedNodeType(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ManagedNodeType", err.Error())
		return
	} else if system {
		resp.Diagnostics.AddError("Invalid ManagedNodeType", "Cannot import resource for system ManagedNodeType")
		return
	} else if data == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		state = *data
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ManagedNodeTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan managedNodeTypeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		description *string
		imageUri    *string
		readme      *string
	)
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.ImageUri.IsNull() || plan.ImageUri.IsUnknown()) {
		imageUri = &plan.ImageUri.Value
	}
	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		readme = &plan.Readme.Value
	}

	echoResp, err := api.UpdateManagedNodeType(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		description,
		imageUri,
		readme,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating ManagedNodeType", err.Error())
		return
	}

	if echoResp.GetManagedNodeType.Update.ConfigTemplate != nil {
		plan.ConfigTemplate = common.Config{Value: *echoResp.GetManagedNodeType.Update.ConfigTemplate}
	} else {
		plan.ConfigTemplate = common.Config{Null: true}
	}
	plan.Description = types.String{Value: echoResp.GetManagedNodeType.Update.Description}
	plan.ImageUri = types.String{Value: echoResp.GetManagedNodeType.Update.ImageUri}
	plan.InUse = types.Bool{Value: echoResp.GetManagedNodeType.Update.InUse}
	plan.MountRequirements = types.Set{ElemType: types.ObjectType{AttrTypes: mountRequirementsAttrTypes()}}
	if len(echoResp.GetManagedNodeType.Update.MountRequirements) > 0 {
		for _, mr := range echoResp.GetManagedNodeType.Update.MountRequirements {
			plan.MountRequirements.Elems = append(
				plan.MountRequirements.Elems,
				types.Object{
					Attrs:     mountRequirementsAttrValues(mr.Description, mr.Source, mr.Target),
					AttrTypes: mountRequirementsAttrTypes(),
				},
			)
		}
	} else {
		plan.MountRequirements.Null = true
	}
	plan.Name = types.String{Value: echoResp.GetManagedNodeType.Update.Name}
	plan.PortRequirements = types.Set{ElemType: types.ObjectType{AttrTypes: portRequirementAttrTypes()}}
	if len(echoResp.GetManagedNodeType.Update.PortRequirements) > 0 {
		for _, pr := range echoResp.GetManagedNodeType.Update.PortRequirements {
			plan.PortRequirements.Elems = append(
				plan.PortRequirements.Elems,
				types.Object{
					Attrs:     portRequirementAttrValues(pr.ContainerPort, pr.Description, pr.Protocol),
					AttrTypes: portRequirementAttrTypes(),
				},
			)
		}
	} else {
		plan.PortRequirements.Null = true
	}
	if echoResp.GetManagedNodeType.Update.Readme != nil {
		plan.Readme = types.String{Value: *echoResp.GetManagedNodeType.Update.Readme}
	} else {
		plan.Readme = types.String{Null: true}
	}
	if echoResp.GetManagedNodeType.Update.ReceiveMessageType != nil {
		plan.ReceiveMessageType = types.String{Value: echoResp.GetManagedNodeType.Update.ReceiveMessageType.Name}
	} else {
		plan.ReceiveMessageType = types.String{Null: true}
	}
	if echoResp.GetManagedNodeType.Update.SendMessageType != nil {
		plan.SendMessageType = types.String{Value: echoResp.GetManagedNodeType.Update.SendMessageType.Name}
	} else {
		plan.SendMessageType = types.String{Null: true}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
