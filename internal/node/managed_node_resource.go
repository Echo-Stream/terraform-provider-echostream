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

type portInputModel struct {
	ContainerPort types.Int64  `tfsdk:"container_port"`
	Description   types.String `tfsdk:"description"`
	HostAddress   types.String `tfsdk:"host_address"`
	HostPort      types.Int64  `tfsdk:"host_port"`
	Protocol      types.String `tfsdk:"protocol"`
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
		s = types.StringValue(*source)
	} else {
		s = types.StringNull()
	}
	return map[string]attr.Value{
		"description": types.StringValue(description),
		"source":      s,
		"target":      types.StringValue(target),
	}
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
		ha = types.StringValue(*hostAddress)
	} else {
		ha = types.StringNull()
	}
	return map[string]attr.Value{
		"container_port": types.Int64Value(int64(contanerPort)),
		"description":    types.StringValue(description),
		"host_address":   ha,
		"host_port":      types.Int64Value(int64(hostPort)),
		"protocol":       types.StringValue(string(protocol)),
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
		diags        diag.Diagnostics
		loggingLevel *api.LogLevel
		mounts       []api.MountInput
		ports        []api.PortInput
	)

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		temp := plan.LoggingLevel.ValueString()
		loggingLevel = (*api.LogLevel)(&temp)
	}
	if !(plan.Mounts.IsNull() || plan.Mounts.IsUnknown()) {
		m := []mountInputModel{}
		resp.Diagnostics.Append(plan.Mounts.ElementsAs(ctx, &m, false)...)
		if !resp.Diagnostics.HasError() && len(m) > 0 {
			mounts = make([]api.MountInput, len(m))
			for i, t_mi := range m {
				mi := api.MountInput{Target: t_mi.Target.ValueString()}
				if !(t_mi.Source.IsNull() || t_mi.Source.IsUnknown()) {
					temp := t_mi.Source.ValueString()
					mi.Source = &temp
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
					ContainerPort: int(t_pi.ContainerPort.ValueInt64()),
					HostPort:      int(t_pi.HostPort.ValueInt64()),
					Protocol:      api.Protocol(t_pi.Protocol.ValueString()),
				}
				if !(t_pi.HostAddress.IsNull() || t_pi.HostAddress.IsUnknown()) {
					temp := t_pi.HostAddress.ValueString()
					pi.HostAddress = &temp
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
		plan.App.ValueString(),
		plan.ManagedNodeType.ValueString(),
		plan.Name.ValueString(),
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
		plan.App = types.StringValue(echoResp.CreateManagedNode.App.Name)
		if echoResp.CreateManagedNode.Config != nil {
			plan.Config = common.ConfigValue(*echoResp.CreateManagedNode.Config)
		} else {
			plan.Config = common.ConfigNull()
		}
		if echoResp.CreateManagedNode.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateManagedNode.Description)
		} else {
			plan.Description = types.StringNull()
		}
		if echoResp.CreateManagedNode.LoggingLevel != nil {
			plan.LoggingLevel = types.StringValue(string(*echoResp.CreateManagedNode.LoggingLevel))
		} else {
			plan.LoggingLevel = types.StringNull()
		}
		plan.ManagedNodeType = types.StringValue(echoResp.CreateManagedNode.ManagedNodeType.Name)
		if len(echoResp.CreateManagedNode.Mounts) > 0 {
			elems := []attr.Value{}
			for _, mount := range echoResp.CreateManagedNode.Mounts {
				if elem, diags := types.ObjectValue(mountAttrTypes(), mountAttrValues(mount.Description, mount.Source, mount.Target)); diags != nil {
					resp.Diagnostics.Append(diags...)
				} else {
					elems = append(elems, elem)
				}
			}
			plan.Mounts, diags = types.SetValue(types.ObjectType{AttrTypes: mountAttrTypes()}, elems)
			if diags != nil && diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
		} else {
			plan.Mounts = types.SetNull(types.ObjectType{AttrTypes: mountAttrTypes()})
		}
		plan.Name = types.StringValue(echoResp.CreateManagedNode.Name)
		if len(echoResp.CreateManagedNode.Ports) > 0 {
			elems := []attr.Value{}
			for _, port := range echoResp.CreateManagedNode.Ports {
				if elem, diags := types.ObjectValue(portAttrTypes(), portAttrValues(port.ContainerPort, port.Description, port.HostAddress, port.HostPort, port.Protocol)); diags != nil {
					resp.Diagnostics.Append(diags...)
				} else {
					elems = append(elems, elem)
				}
			}
			plan.Ports, diags = types.SetValue(types.ObjectType{AttrTypes: portAttrTypes()}, elems)
			if diags != nil && diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
		} else {
			plan.Ports = types.SetNull(types.ObjectType{AttrTypes: portAttrTypes()})
		}
		if echoResp.CreateManagedNode.ReceiveMessageType != nil {
			plan.ReceiveMessageType = types.StringValue(echoResp.CreateManagedNode.ReceiveMessageType.Name)
		} else {
			plan.ReceiveMessageType = types.StringNull()
		}
		if echoResp.CreateManagedNode.SendMessageType != nil {
			plan.SendMessageType = types.StringValue(echoResp.CreateManagedNode.SendMessageType.Name)
		} else {
			plan.SendMessageType = types.StringNull()
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

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
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
				MarkdownDescription: "The ManagedApp that this Node is associated with.",
				Required:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Type:                types.StringType,
			},
			"config": {
				MarkdownDescription: "The config, in JSON object format (i.e. - dict, map).",
				Optional:            true,
				Sensitive:           true,
				Type:                common.ConfigType{},
			},
			"logging_level": {
				MarkdownDescription: "The logging level. One of `DEBUG`, `ERROR`, `INFO`, `WARNING`. Defaults to `INFO`.",
				Optional:            true,
				Type:                types.StringType,
				Validators:          []tfsdk.AttributeValidator{common.LogLevelValidator},
			},
			"managed_node_type": {
				MarkdownDescription: "The ManagedNodeType of this ManagedNode. This Node must conform to all of the" +
					" config, mount and port requirements specified in the ManagedNodeType.",
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Required:      true,
				Type:          types.StringType,
			},
			"mounts": {
				Attributes: tfsdk.SetNestedAttributes(
					map[string]tfsdk.Attribute{
						"description": {
							Computed:            true,
							MarkdownDescription: " A human-readable description.",
							Type:                types.StringType,
						},
						"source": {
							MarkdownDescription: "The source of the mount. If not present, an anonymous volume will be created.",
							Optional:            true,
							Type:                types.StringType,
						},
						"target": {
							MarkdownDescription: "The path to mount the volume in the Docker container.",
							Required:            true,
							Type:                types.StringType,
						},
					},
				),
				MarkdownDescription: "A list of the mounts (i.e. - volumes) used by the Docker container.",
				Optional:            true,
			},
			"ports": {
				Attributes: tfsdk.SetNestedAttributes(
					map[string]tfsdk.Attribute{
						"container_port": {
							MarkdownDescription: "The exposed container port.",
							Required:            true,
							Type:                types.Int64Type,
						},
						"description": {
							Computed:            true,
							MarkdownDescription: "A human-readable description.",
							Type:                types.StringType,
						},
						"host_address": {
							MarkdownDescription: "The host address the port is exposed on. Defaults to `0.0.0.0`.",
							Optional:            true,
							Type:                types.StringType,
							Validators:          []tfsdk.AttributeValidator{validators.Ipaddr()},
						},
						"host_port": {
							MarkdownDescription: "The exposed host port. Must be between `1024` and `65535`, inclusive.",
							Required:            true,
							Type:                types.Int64Type,
							Validators:          []tfsdk.AttributeValidator{common.PortValidator},
						},
						"protocol": {
							MarkdownDescription: "The protocol to use for the port. One of `sctp`, `tcp` or `udp`.",
							Required:            true,
							Type:                types.StringType,
							Validators:          []tfsdk.AttributeValidator{common.ProtocolValidator},
						},
					},
				),
				MarkdownDescription: "A list of ports exposed by the Docker container.",
				Optional:            true,
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		MarkdownDescription: "[ManagedNodes](https://docs.echo.stream/docs/managed-node) are instances of Docker containers that exist within ManagedApps.",
	}, nil
}

func (r *ManagedNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ManagedNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_node"
}

func (r *ManagedNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		state managedNodeModel
	)

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ManagedNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeManagedNode:
			state.App = types.StringValue(node.App.Name)
			if node.Config != nil {
				state.Config = common.ConfigValue(*node.Config)
			} else {
				state.Config = common.ConfigNull()
			}
			if node.Description != nil {
				state.Description = types.StringValue(*node.Description)
			} else {
				state.Description = types.StringNull()
			}
			if node.LoggingLevel != nil {
				state.LoggingLevel = types.StringValue(string(*node.LoggingLevel))
			} else {
				state.LoggingLevel = types.StringNull()
			}
			state.ManagedNodeType = types.StringValue(node.ManagedNodeType.Name)
			if len(node.Mounts) > 0 {
				elems := []attr.Value{}
				for _, mount := range node.Mounts {
					if elem, diags := types.ObjectValue(mountAttrTypes(), mountAttrValues(mount.Description, mount.Source, mount.Target)); diags != nil {
						resp.Diagnostics.Append(diags...)
					} else {
						elems = append(elems, elem)
					}
				}
				state.Mounts, diags = types.SetValue(types.ObjectType{AttrTypes: mountAttrTypes()}, elems)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			} else {
				state.Mounts = types.SetNull(types.ObjectType{AttrTypes: mountAttrTypes()})
			}
			state.Name = types.StringValue(node.Name)
			if len(node.Ports) > 0 {
				elems := []attr.Value{}
				for _, port := range node.Ports {
					if elem, diags := types.ObjectValue(portAttrTypes(), portAttrValues(port.ContainerPort, port.Description, port.HostAddress, port.HostPort, port.Protocol)); diags != nil {
						resp.Diagnostics.Append(diags...)
					} else {
						elems = append(elems, elem)
					}
				}
				state.Ports, diags = types.SetValue(types.ObjectType{AttrTypes: portAttrTypes()}, elems)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			} else {
				state.Ports = types.SetNull(types.ObjectType{AttrTypes: portAttrTypes()})
			}
			if node.ReceiveMessageType != nil {
				state.ReceiveMessageType = types.StringValue(node.ReceiveMessageType.Name)
			} else {
				state.ReceiveMessageType = types.StringNull()
			}
			if node.SendMessageType != nil {
				state.SendMessageType = types.StringValue(node.SendMessageType.Name)
			} else {
				state.SendMessageType = types.StringNull()
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ManagedNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.ValueString()),
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
		diags        diag.Diagnostics
		loggingLevel *api.LogLevel
		mounts       []api.MountInput
		ports        []api.PortInput
	)
	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		temp := plan.LoggingLevel.ValueString()
		loggingLevel = (*api.LogLevel)(&temp)
	}
	if !(plan.Mounts.IsNull() || plan.Mounts.IsUnknown()) {
		m := []mountInputModel{}
		resp.Diagnostics.Append(plan.Mounts.ElementsAs(ctx, &m, false)...)
		if !resp.Diagnostics.HasError() && len(m) > 0 {
			mounts = make([]api.MountInput, len(m))
			for i, t_mi := range m {
				mi := api.MountInput{Target: t_mi.Target.ValueString()}
				if !(t_mi.Source.IsNull() || t_mi.Source.IsUnknown()) {
					temp := t_mi.Source.ValueString()
					mi.Source = &temp
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
					ContainerPort: int(t_pi.ContainerPort.ValueInt64()),
					HostPort:      int(t_pi.HostPort.ValueInt64()),
					Protocol:      api.Protocol(t_pi.Protocol.ValueString()),
				}
				if !(t_pi.HostAddress.IsNull() || t_pi.HostAddress.IsUnknown()) {
					temp := t_pi.HostAddress.ValueString()
					pi.HostAddress = &temp
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
		plan.Name.ValueString(),
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
		resp.Diagnostics.AddError("Cannot find ManagedNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.ValueString()))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateManagedNodeGetNodeManagedNode:
			plan.App = types.StringValue(node.Update.App.Name)
			if node.Update.Config != nil {
				plan.Config = common.ConfigValue(*node.Update.Config)
			} else {
				plan.Config = common.ConfigNull()
			}
			if node.Update.Description != nil {
				plan.Description = types.StringValue(*node.Update.Description)
			} else {
				plan.Description = types.StringNull()
			}
			if node.Update.LoggingLevel != nil {
				plan.LoggingLevel = types.StringValue(string(*node.Update.LoggingLevel))
			} else {
				plan.LoggingLevel = types.StringNull()
			}
			plan.ManagedNodeType = types.StringValue(node.Update.ManagedNodeType.Name)
			if len(node.Update.Mounts) > 0 {
				elems := []attr.Value{}
				for _, mount := range node.Update.Mounts {
					if elem, diags := types.ObjectValue(mountAttrTypes(), mountAttrValues(mount.Description, mount.Source, mount.Target)); diags != nil {
						resp.Diagnostics.Append(diags...)
					} else {
						elems = append(elems, elem)
					}
				}
				plan.Mounts, diags = types.SetValue(types.ObjectType{AttrTypes: mountAttrTypes()}, elems)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			} else {
				plan.Mounts = types.SetNull(types.ObjectType{AttrTypes: mountAttrTypes()})
			}
			plan.Name = types.StringValue(node.Update.Name)
			if len(node.Update.Ports) > 0 {
				elems := []attr.Value{}
				for _, port := range node.Update.Ports {
					if elem, diags := types.ObjectValue(portAttrTypes(), portAttrValues(port.ContainerPort, port.Description, port.HostAddress, port.HostPort, port.Protocol)); diags != nil {
						resp.Diagnostics.Append(diags...)
					} else {
						elems = append(elems, elem)
					}
				}
				plan.Ports, diags = types.SetValue(types.ObjectType{AttrTypes: portAttrTypes()}, elems)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			} else {
				plan.Ports = types.SetNull(types.ObjectType{AttrTypes: portAttrTypes()})
			}
			if node.Update.ReceiveMessageType != nil {
				plan.ReceiveMessageType = types.StringValue(node.Update.ReceiveMessageType.Name)
			} else {
				plan.ReceiveMessageType = types.StringNull()
			}
			if node.Update.SendMessageType != nil {
				plan.SendMessageType = types.StringValue(node.Update.SendMessageType.Name)
			} else {
				plan.SendMessageType = types.StringNull()
			}
		default:
			resp.Diagnostics.AddError(
				"Expected ManagedNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
