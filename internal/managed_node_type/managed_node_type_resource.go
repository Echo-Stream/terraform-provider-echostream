package managed_node_type

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithImportState = &ManagedNodeTypeResource{}
	_ resource.ResourceWithModifyPlan  = &ManagedNodeTypeResource{}
	_ resource.ResourceWithSchema      = &ManagedNodeTypeResource{}
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
		diags              diag.Diagnostics
		mountRequirements  []api.MountRequirementInput
		portRequirements   []api.PortRequirementInput
		readme             *string
		receiveMessageType *string
		sendMessageType    *string
	)

	if !(plan.ConfigTemplate.IsNull() || plan.ConfigTemplate.IsUnknown()) {
		temp := plan.ConfigTemplate.ValueConfig()
		configTemplate = &temp
	}
	if !(plan.MountRequirements.IsNull() || plan.MountRequirements.IsUnknown()) {
		mr := []mountRequirementsModel{}
		resp.Diagnostics.Append(plan.MountRequirements.ElementsAs(ctx, &mr, false)...)
		if !resp.Diagnostics.HasError() {
			mountRequirements = make([]api.MountRequirementInput, len(mr))
			for i, t_mri := range mr {
				mri := api.MountRequirementInput{
					Description: t_mri.Description.ValueString(),
					Target:      t_mri.Target.ValueString(),
				}
				if !(t_mri.Source.IsNull() || t_mri.Source.IsUnknown()) {
					temp := t_mri.Source.ValueString()
					mri.Source = &temp
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
					ContainerPort: int(t_pri.ContainerPort.ValueInt64()),
					Description:   t_pri.Description.ValueString(),
					Protocol:      api.Protocol(t_pri.Protocol.ValueString()),
				}
				portRequirements[i] = pri
			}
		}
	}
	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		temp := plan.Readme.ValueString()
		readme = &temp
	}
	if !(plan.ReceiveMessageType.IsNull() || plan.ReceiveMessageType.IsUnknown()) {
		temp := plan.ReceiveMessageType.ValueString()
		receiveMessageType = &temp
	}
	if !(plan.SendMessageType.IsNull() || plan.SendMessageType.IsUnknown()) {
		temp := plan.SendMessageType.ValueString()
		sendMessageType = &temp
	}

	if resp.Diagnostics.HasError() {
		return
	}

	echoResp, err := api.CreateManagedNodeType(
		ctx,
		r.data.Client,
		plan.Description.ValueString(),
		plan.ImageUri.ValueString(),
		plan.Name.ValueString(),
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
		plan.ConfigTemplate = common.ConfigValue(*echoResp.CreateManagedNodeType.ConfigTemplate)
	} else {
		plan.ConfigTemplate = common.ConfigNull()
	}
	plan.Description = types.StringValue(echoResp.CreateManagedNodeType.Description)
	plan.Id = types.StringValue(echoResp.CreateManagedNodeType.Name)
	plan.ImageUri = types.StringValue(echoResp.CreateManagedNodeType.ImageUri)
	plan.InUse = types.BoolValue(echoResp.CreateManagedNodeType.InUse)
	if len(echoResp.CreateManagedNodeType.MountRequirements) > 0 {
		elems := []attr.Value{}
		for _, mountReq := range echoResp.CreateManagedNodeType.MountRequirements {
			if elem, d := types.ObjectValue(mountRequirementsAttrTypes(), mountRequirementsAttrValues(mountReq.Description, mountReq.Source, mountReq.Target)); d != nil {
				resp.Diagnostics.Append(d...)
			} else {
				elems = append(elems, elem)
			}
		}
		plan.MountRequirements, diags = types.SetValue(types.ObjectType{AttrTypes: mountRequirementsAttrTypes()}, elems)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
	} else {
		plan.MountRequirements = types.SetNull(types.ObjectType{AttrTypes: mountRequirementsAttrTypes()})
	}
	plan.Name = types.StringValue(echoResp.CreateManagedNodeType.Name)
	if len(echoResp.CreateManagedNodeType.PortRequirements) > 0 {
		elems := []attr.Value{}
		for _, portReq := range echoResp.CreateManagedNodeType.PortRequirements {
			if elem, d := types.ObjectValue(portRequirementAttrTypes(), portRequirementAttrValues(portReq.ContainerPort, portReq.Description, portReq.Protocol)); d != nil {
				resp.Diagnostics.Append(d...)
			} else {
				elems = append(elems, elem)
			}
		}
		plan.PortRequirements, diags = types.SetValue(types.ObjectType{AttrTypes: portRequirementAttrTypes()}, elems)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
	} else {
		plan.PortRequirements = types.SetNull(types.ObjectType{AttrTypes: portRequirementAttrTypes()})
	}
	if echoResp.CreateManagedNodeType.Readme != nil {
		plan.Readme = types.StringValue(*echoResp.CreateManagedNodeType.Readme)
	} else {
		plan.Readme = types.StringNull()
	}
	if echoResp.CreateManagedNodeType.ReceiveMessageType != nil {
		plan.ReceiveMessageType = types.StringValue(echoResp.CreateManagedNodeType.ReceiveMessageType.Name)
	} else {
		plan.ReceiveMessageType = types.StringNull()
	}
	if echoResp.CreateManagedNodeType.SendMessageType != nil {
		plan.SendMessageType = types.StringValue(echoResp.CreateManagedNodeType.SendMessageType.Name)
	} else {
		plan.SendMessageType = types.StringNull()
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

	if _, err := api.DeleteManagedNodeType(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ManagedNodeType", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
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
	if !state.InUse.ValueBool() {
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
		resp.Diagnostics.AddError("Cannot destroy ManagedNodeType", fmt.Sprintf("ManagedNodeType %s is in use and may not be destroyed", state.Name.ValueString()))
	}
}

func (r *ManagedNodeTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state managedNodeTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, system, diags := readManagedNodeType(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); diags != nil {
		resp.Diagnostics.Append(diags...)
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

func (r *ManagedNodeTypeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config_template": schema.StringAttribute{
				CustomType: common.ConfigType{},
				MarkdownDescription: "A [JSON Schema](https://json-schema.org/) document that specifies the" +
					" requirements for the config attribute of ManagedNodes created using this ManagedNodeType.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"image_uri": schema.StringAttribute{
				MarkdownDescription: "The URI of the Docker image. Must be a [public](https://docs.aws.amazon.com/AmazonECR/latest/public/public-repositories.html) " +
					"or a [private](https://docs.aws.amazon.com/AmazonECR/latest/userguide/Repositories.html) AWS ECR repository.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(?:(?:[0-9]{12}\.dkr\.ecr\.[a-z]+\-[a-z]+\-[0-9]\.amazonaws\.com/.+\:.+)|(?:public\.ecr\.aws/.+/.+\:.+))$`),
						`must be either a private ECR image URI (aws_account_id.dkr.ecr.region.amazonaws.com/respository:tag) or a public ECR image URI (public.ecr.aws/registry_alias/repository:tag)`,
					),
				},
			},
			"in_use": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: " True if this is used by ManagedNodes.",
			},
			"mount_requirements": schema.SetNestedAttribute{
				MarkdownDescription: "The mount (i.e. - volume) requirements of the Docker image.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							MarkdownDescription: "A human-readable description of the port.",
							Required:            true,
						},
						"source": schema.StringAttribute{
							MarkdownDescription: "The path of the mount on the host.",
							Optional:            true,
						},
						"target": schema.StringAttribute{
							MarkdownDescription: "The path of the mount in the Docker container.",
							Required:            true,
						},
					},
				},
				Optional:      true,
				PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the ManagedNodeType. Must be unique within the Tenant.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
				Validators:          append(common.NameValidators, common.NotSystemNameValidator),
			},
			"port_requirements": schema.SetNestedAttribute{
				MarkdownDescription: "The port requirements of the Docker image.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"container_port": schema.Int64Attribute{
							MarkdownDescription: "The exposed container port.",
							Required:            true,
							Validators:          []validator.Int64{int64validator.Between(1, 65535)},
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "A human-readable description for the port.",
							Required:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "The protocol to use for the port. One of `sctp`, `tcp` or `udp`.",
							Required:            true,
							Validators:          []validator.String{common.ProtocolValidator},
						},
					},
				},
				Optional:      true,
				PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()},
			},
			"readme": schema.StringAttribute{
				MarkdownDescription: "README in MarkDown format.",
				Optional:            true,
			},
			"receive_message_type": schema.StringAttribute{
				MarkdownDescription: "The MessageType that ManagedNodes created with this ManagedNodeType are capable of receiving.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"send_message_type": schema.StringAttribute{
				MarkdownDescription: "The MessageType that ManagedNodes created with this ManagedNodeType are capable of sending.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
		MarkdownDescription: "ManagedNodeTypes are wrappers around Docker image definitions and define the requirements " +
			"necessary to instantiate those images as Docker containers inside a ManagedNode.",
	}
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
		diags       diag.Diagnostics
		imageUri    *string
		readme      *string
	)
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.ImageUri.IsNull() || plan.ImageUri.IsUnknown()) {
		temp := plan.ImageUri.ValueString()
		imageUri = &temp
	}
	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		temp := plan.Readme.ValueString()
		readme = &temp
	}

	echoResp, err := api.UpdateManagedNodeType(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
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
		plan.ConfigTemplate = common.ConfigValue(*echoResp.GetManagedNodeType.Update.ConfigTemplate)
	} else {
		plan.ConfigTemplate = common.ConfigNull()
	}
	plan.Description = types.StringValue(echoResp.GetManagedNodeType.Update.Description)
	plan.Id = types.StringValue(echoResp.GetManagedNodeType.Update.Name)
	plan.ImageUri = types.StringValue(echoResp.GetManagedNodeType.Update.ImageUri)
	plan.InUse = types.BoolValue(echoResp.GetManagedNodeType.Update.InUse)
	if len(echoResp.GetManagedNodeType.Update.MountRequirements) > 0 {
		elems := []attr.Value{}
		for _, mountReq := range echoResp.GetManagedNodeType.Update.MountRequirements {
			if elem, d := types.ObjectValue(mountRequirementsAttrTypes(), mountRequirementsAttrValues(mountReq.Description, mountReq.Source, mountReq.Target)); d != nil {
				resp.Diagnostics.Append(d...)
			} else {
				elems = append(elems, elem)
			}
		}
		plan.MountRequirements, diags = types.SetValue(types.ObjectType{AttrTypes: mountRequirementsAttrTypes()}, elems)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
	} else {
		plan.MountRequirements = types.SetNull(types.ObjectType{AttrTypes: mountRequirementsAttrTypes()})
	}
	plan.Name = types.StringValue(echoResp.GetManagedNodeType.Update.Name)
	if len(echoResp.GetManagedNodeType.Update.PortRequirements) > 0 {
		elems := []attr.Value{}
		for _, portReq := range echoResp.GetManagedNodeType.Update.PortRequirements {
			if elem, d := types.ObjectValue(portRequirementAttrTypes(), portRequirementAttrValues(portReq.ContainerPort, portReq.Description, portReq.Protocol)); d != nil {
				resp.Diagnostics.Append(d...)
			} else {
				elems = append(elems, elem)
			}
		}
		plan.PortRequirements, diags = types.SetValue(types.ObjectType{AttrTypes: portRequirementAttrTypes()}, elems)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
	} else {
		plan.PortRequirements = types.SetNull(types.ObjectType{AttrTypes: portRequirementAttrTypes()})
	}
	if echoResp.GetManagedNodeType.Update.Readme != nil {
		plan.Readme = types.StringValue(*echoResp.GetManagedNodeType.Update.Readme)
	} else {
		plan.Readme = types.StringNull()
	}
	if echoResp.GetManagedNodeType.Update.ReceiveMessageType != nil {
		plan.ReceiveMessageType = types.StringValue(echoResp.GetManagedNodeType.Update.ReceiveMessageType.Name)
	} else {
		plan.ReceiveMessageType = types.StringNull()
	}
	if echoResp.GetManagedNodeType.Update.SendMessageType != nil {
		plan.SendMessageType = types.StringValue(echoResp.GetManagedNodeType.Update.SendMessageType.Name)
	} else {
		plan.SendMessageType = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
