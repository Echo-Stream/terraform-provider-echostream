package node

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.ResourceWithImportState = &ManagedNodeResource{}

// ManagedNodeResource defines the resource implementation.
type ManagedNodeResource struct {
	data *common.ProviderData
}

type managedNodeModel struct {
	App                types.String  `tfsdk:"app"`
	Config             common.Config `tfsdk:"config"`
	Description        types.String  `tfsdk:"description"`
	LoggingLevel       types.String  `tfsdk:"logging_level"`
	ManagedNodeType    types.String  `tfsdk:"managed_node_type"`
	Mounts             types.Set     `tfsdk:"mounts"`
	Name               types.String  `tfsdk:"name"`
	Ports              types.Set     `tfsdk:"ports"`
	ReceiveMessageType types.String  `tfsdk:"receive_message_type"`
	SendMessageType    types.String  `tfsdk:"send_message_type"`
}

type mountInputModel struct {
	Description types.String `tfsdk:"description"`
	Source      types.String `tfsdk:"source"`
	Target      types.String `tfsdk:"target"`
}

func mountAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"description": types.StringType,
		"source":      types.StringType,
		"target":      types.StringType,
	}
}

func mountAttrValues(
	description string,
	source *string,
	target string,
) map[string]attr.Value {
	var s types.String
	if source != nil {
		s = types.String{Value: *source}
	} else {
		s = types.String{Null: true}
	}
	return map[string]attr.Value{
		"description": types.String{Value: description},
		"source":      s,
		"target":      types.String{Value: target},
	}
}

type portInputModel struct {
	ContainerPort types.Int64  `tfsdk:"container_port"`
	Description   types.String `tfsdk:"description"`
	HostAddress   types.String `tfsdk:"host_address"`
	HostPort      types.Int64  `tfsdk:"host_port"`
	Protocol      types.String `tfsdk:"protocol"`
}

func portAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"container_port": types.Int64Type,
		"description":    types.StringType,
		"host_address":   types.StringType,
		"host_port":      types.Int64Type,
		"protocol":       types.StringType,
	}
}

func portAttrValues(
	contanerPort int,
	description string,
	hostAddress *string,
	hostPort int,
	protocol api.Protocol,
) map[string]attr.Value {
	var ha types.String
	if hostAddress != nil {
		ha = types.String{Value: *hostAddress}
	} else {
		ha = types.String{Null: true}
	}
	return map[string]attr.Value{
		"container_port": types.Int64{Value: int64(contanerPort)},
		"description":    types.String{Value: description},
		"host_address":   ha,
		"host_port":      types.Int64{Value: int64(hostPort)},
		"protocol":       types.String{Value: string(protocol)},
	}
}

func (r *ManagedNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ManagedNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan managedNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config       *string
		description  *string
		loggingLevel *api.LogLevel
		mounts       []api.MountInput
		ports        []api.PortInput
	)

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		loggingLevel = (*api.LogLevel)(&plan.LoggingLevel.Value)
	}
	if !(plan.Mounts.IsNull() || plan.Mounts.IsUnknown()) {
		m := []mountInputModel{}
		resp.Diagnostics.Append(plan.Mounts.ElementsAs(ctx, &m, false)...)
		if !resp.Diagnostics.HasError() && len(m) > 0 {
			mounts = make([]api.MountInput, len(m))
			for i, t_mi := range m {
				mi := api.MountInput{Target: t_mi.Target.Value}
				if !(t_mi.Source.IsNull() || t_mi.Source.IsUnknown()) {
					mi.Source = &t_mi.Source.Value
				}
				mounts[i] = mi
			}
		}
	}
	if !(plan.Ports.IsNull() || plan.Ports.IsUnknown()) {
		p := []portInputModel{}
		resp.Diagnostics.Append(plan.Ports.ElementsAs(ctx, &p, false)...)
		if !resp.Diagnostics.HasError() && len(p) > 0 {
			ports = make([]api.PortInput, len(p))
			for i, t_pi := range p {
				pi := api.PortInput{
					ContainerPort: int(t_pi.ContainerPort.Value),
					HostPort:      int(t_pi.HostPort.Value),
					Protocol:      api.Protocol(t_pi.Protocol.Value),
				}
				if !(t_pi.HostAddress.IsNull() || t_pi.HostAddress.IsUnknown()) {
					pi.HostAddress = &t_pi.HostAddress.Value
				}
				ports[i] = pi
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.CreateManagedNode(
		ctx,
		r.data.Client,
		plan.App.Value,
		plan.ManagedNodeType.Value,
		plan.Name.Value,
		r.data.Tenant,
		config,
		description,
		loggingLevel,
		mounts,
		ports,
	); err != nil {
		resp.Diagnostics.AddError("Error creating ManagedNode", err.Error())
		return
	} else {
		plan.App = types.String{Value: echoResp.CreateManagedNode.App.Name}
		if echoResp.CreateManagedNode.Config != nil {
			plan.Config = common.Config{Value: *echoResp.CreateManagedNode.Config}
		} else {
			plan.Config = common.Config{Null: true}
		}
		if echoResp.CreateManagedNode.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateManagedNode.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		if echoResp.CreateManagedNode.LoggingLevel != nil {
			plan.LoggingLevel = types.String{Value: string(*echoResp.CreateManagedNode.LoggingLevel)}
		} else {
			plan.LoggingLevel = types.String{Null: true}
		}
		plan.ManagedNodeType = types.String{Value: echoResp.CreateManagedNode.ManagedNodeType.Name}
		plan.Mounts = types.Set{ElemType: types.ObjectType{AttrTypes: mountAttrTypes()}}
		if len(echoResp.CreateManagedNode.Mounts) > 0 {
			for _, mount := range echoResp.CreateManagedNode.Mounts {
				plan.Mounts.Elems = append(
					plan.Mounts.Elems,
					types.Object{
						Attrs:     mountAttrValues(mount.Description, mount.Source, mount.Target),
						AttrTypes: mountAttrTypes(),
					},
				)
			}
		} else {
			plan.Mounts.Null = true
		}
		plan.Name = types.String{Value: echoResp.CreateManagedNode.Name}
		plan.Ports = types.Set{ElemType: types.ObjectType{AttrTypes: portAttrTypes()}}
		if len(echoResp.CreateManagedNode.Ports) > 0 {
			for _, port := range echoResp.CreateManagedNode.Ports {
				plan.Ports.Elems = append(
					plan.Ports.Elems,
					types.Object{
						Attrs: portAttrValues(
							port.ContainerPort,
							port.Description,
							port.HostAddress,
							port.HostPort,
							port.Protocol,
						),
						AttrTypes: portAttrTypes(),
					},
				)
			}
		} else {
			plan.Ports.Null = true
		}
		if echoResp.CreateManagedNode.ReceiveMessageType != nil {
			plan.ReceiveMessageType = types.String{Value: echoResp.CreateManagedNode.ReceiveMessageType.Name}
		} else {
			plan.ReceiveMessageType = types.String{Null: true}
		}
		if echoResp.CreateManagedNode.SendMessageType != nil {
			plan.SendMessageType = types.String{Value: echoResp.CreateManagedNode.SendMessageType.Name}
		} else {
			plan.SendMessageType = types.String{Null: true}
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ManagedNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state managedNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting ManagedNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *ManagedNodeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := dataSendReceiveNodeSchema()
	for key, attribute := range schema {
		switch key {
		case "description":
			attribute.Computed = false
			attribute.Optional = true
		case "name":
			attribute.Computed = false
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
			attribute.Required = true
			attribute.Validators = []tfsdk.AttributeValidator{
				stringvalidator.LengthBetween(3, 63),
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^[A-Za-z0-9\-]*$`),
					"value must contain only lowercase/uppercase alphanumeric characters, or \"-\"",
				),
			}
		}
		schema[key] = attribute
	}
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"app": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Type:                types.StringType,
			},
			"config": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Sensitive:           true,
				Type:                common.ConfigType{},
			},
			"logging_level": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
				Validators:          []tfsdk.AttributeValidator{common.LogLevelValidator},
			},
			"managed_node_type": {
				Description:         "",
				MarkdownDescription: "",
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Required:            true,
				Type:                types.StringType,
			},
			"mounts": {
				Attributes: tfsdk.SetNestedAttributes(
					map[string]tfsdk.Attribute{
						"description": {
							Computed:            true,
							Description:         "",
							MarkdownDescription: "",
							Type:                types.StringType,
						},
						"source": {
							Description:         "",
							MarkdownDescription: "",
							Optional:            true,
							Type:                types.StringType,
						},
						"target": {
							Description:         "",
							MarkdownDescription: "",
							Required:            true,
							Type:                types.StringType,
						},
					},
				),
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
			},
			"ports": {
				Attributes: tfsdk.SetNestedAttributes(
					map[string]tfsdk.Attribute{
						"container_port": {
							Description:         "",
							MarkdownDescription: "",
							Required:            true,
							Type:                types.Int64Type,
							Validators:          []tfsdk.AttributeValidator{common.PortValidator},
						},
						"description": {
							Computed:            true,
							Description:         "",
							MarkdownDescription: "",
							Type:                types.StringType,
						},
						"host_address": {
							Description:         "",
							MarkdownDescription: "",
							Optional:            true,
							Type:                types.StringType,
							Validators:          []tfsdk.AttributeValidator{validators.Ipaddr()},
						},
						"host_port": {
							Description:         "",
							MarkdownDescription: "",
							Required:            true,
							Type:                types.Int64Type,
							Validators:          []tfsdk.AttributeValidator{common.PortValidator},
						},
						"protocol": {
							Description:         "",
							MarkdownDescription: "",
							Required:            true,
							Type:                types.StringType,
							Validators:          []tfsdk.AttributeValidator{common.ProtocolValidator},
						},
					},
				),
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		Description:         "ManagedNodes run inside of ManagedApps",
		MarkdownDescription: "ManagedNodes run inside of ManagedApps",
	}, nil
}

func (r *ManagedNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ManagedNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_node"
}

func (r *ManagedNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state managedNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ManagedNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeManagedNode:
			state.App = types.String{Value: node.App.Name}
			if node.Config != nil {
				state.Config = common.Config{Value: *node.Config}
			} else {
				state.Config = common.Config{Null: true}
			}
			if node.Description != nil {
				state.Description = types.String{Value: *node.Description}
			} else {
				state.Description = types.String{Null: true}
			}
			if node.LoggingLevel != nil {
				state.LoggingLevel = types.String{Value: string(*node.LoggingLevel)}
			} else {
				state.LoggingLevel = types.String{Null: true}
			}
			state.ManagedNodeType = types.String{Value: node.ManagedNodeType.Name}
			state.Mounts = types.Set{ElemType: types.ObjectType{AttrTypes: mountAttrTypes()}}
			if len(node.Mounts) > 0 {
				for _, mount := range node.Mounts {
					state.Mounts.Elems = append(
						state.Mounts.Elems,
						types.Object{
							Attrs:     mountAttrValues(mount.Description, mount.Source, mount.Target),
							AttrTypes: mountAttrTypes(),
						},
					)
				}
			} else {
				state.Mounts.Null = true
			}
			state.Name = types.String{Value: node.Name}
			state.Ports = types.Set{ElemType: types.ObjectType{AttrTypes: portAttrTypes()}}
			if len(node.Ports) > 0 {
				for _, port := range node.Ports {
					state.Ports.Elems = append(
						state.Ports.Elems,
						types.Object{
							Attrs: portAttrValues(
								port.ContainerPort,
								port.Description,
								port.HostAddress,
								port.HostPort,
								port.Protocol,
							),
							AttrTypes: portAttrTypes(),
						},
					)
				}
			} else {
				state.Ports.Null = true
			}
			if node.ReceiveMessageType != nil {
				state.ReceiveMessageType = types.String{Value: node.ReceiveMessageType.Name}
			} else {
				state.ReceiveMessageType = types.String{Null: true}
			}
			if node.SendMessageType != nil {
				state.SendMessageType = types.String{Value: node.SendMessageType.Name}
			} else {
				state.SendMessageType = types.String{Null: true}
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ManagedNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ManagedNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan managedNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config       *string
		description  *string
		loggingLevel *api.LogLevel
		mounts       []api.MountInput
		ports        []api.PortInput
	)
	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		loggingLevel = (*api.LogLevel)(&plan.LoggingLevel.Value)
	}
	if !(plan.Mounts.IsNull() || plan.Mounts.IsUnknown()) {
		m := []mountInputModel{}
		resp.Diagnostics.Append(plan.Mounts.ElementsAs(ctx, &m, false)...)
		if !resp.Diagnostics.HasError() && len(m) > 0 {
			mounts = make([]api.MountInput, len(m))
			for i, t_mi := range m {
				mi := api.MountInput{Target: t_mi.Target.Value}
				if !(t_mi.Source.IsNull() || t_mi.Source.IsUnknown()) {
					mi.Source = &t_mi.Source.Value
				}
				mounts[i] = mi
			}
		}
	}
	if !(plan.Ports.IsNull() || plan.Ports.IsUnknown()) {
		p := []portInputModel{}
		resp.Diagnostics.Append(plan.Ports.ElementsAs(ctx, &p, false)...)
		if !resp.Diagnostics.HasError() && len(p) > 0 {
			ports = make([]api.PortInput, len(p))
			for i, t_pi := range p {
				pi := api.PortInput{
					ContainerPort: int(t_pi.ContainerPort.Value),
					HostPort:      int(t_pi.HostPort.Value),
					Protocol:      api.Protocol(t_pi.Protocol.Value),
				}
				if !(t_pi.HostAddress.IsNull() || t_pi.HostAddress.IsUnknown()) {
					pi.HostAddress = &t_pi.HostAddress.Value
				}
				ports[i] = pi
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.UpdateManagedNode(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		config,
		description,
		loggingLevel,
		mounts,
		ports,
	); err != nil {
		resp.Diagnostics.AddError("Error updating ManagedNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find ManagedNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.Value))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateManagedNodeGetNodeManagedNode:
			plan.App = types.String{Value: node.Update.App.Name}
			if node.Update.Config != nil {
				plan.Config = common.Config{Value: *node.Update.Config}
			} else {
				plan.Config = common.Config{Null: true}
			}
			if node.Update.Description != nil {
				plan.Description = types.String{Value: *node.Update.Description}
			} else {
				plan.Description = types.String{Null: true}
			}
			if node.Update.LoggingLevel != nil {
				plan.LoggingLevel = types.String{Value: string(*node.Update.LoggingLevel)}
			} else {
				plan.LoggingLevel = types.String{Null: true}
			}
			plan.ManagedNodeType = types.String{Value: node.Update.ManagedNodeType.Name}
			plan.Mounts = types.Set{ElemType: types.ObjectType{AttrTypes: mountAttrTypes()}}
			if len(node.Update.Mounts) > 0 {
				for _, mount := range node.Update.Mounts {
					plan.Mounts.Elems = append(
						plan.Mounts.Elems,
						types.Object{
							Attrs:     mountAttrValues(mount.Description, mount.Source, mount.Target),
							AttrTypes: mountAttrTypes(),
						},
					)
				}
			} else {
				plan.Mounts.Null = true
			}
			plan.Name = types.String{Value: node.Update.Name}
			plan.Ports = types.Set{ElemType: types.ObjectType{AttrTypes: portAttrTypes()}}
			if len(node.Update.Ports) > 0 {
				for _, port := range node.Update.Ports {
					plan.Ports.Elems = append(
						plan.Ports.Elems,
						types.Object{
							Attrs: portAttrValues(
								port.ContainerPort,
								port.Description,
								port.HostAddress,
								port.HostPort,
								port.Protocol,
							),
							AttrTypes: portAttrTypes(),
						},
					)
				}
			} else {
				plan.Ports.Null = true
			}
			if node.Update.ReceiveMessageType != nil {
				plan.ReceiveMessageType = types.String{Value: node.Update.ReceiveMessageType.Name}
			} else {
				plan.ReceiveMessageType = types.String{Null: true}
			}
			if node.Update.SendMessageType != nil {
				plan.SendMessageType = types.String{Value: node.Update.SendMessageType.Name}
			} else {
				plan.SendMessageType = types.String{Null: true}
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ManagedNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
