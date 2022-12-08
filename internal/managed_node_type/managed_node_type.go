package managed_node_type

import (
	"context"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type managedNodeTypeModel struct {
	ConfigTemplate     common.Config `tfsdk:"config_template"`
	Description        types.String  `tfsdk:"description"`
	Id                 types.String  `tfsdk:"id"`
	ImageUri           types.String  `tfsdk:"image_uri"`
	InUse              types.Bool    `tfsdk:"in_use"`
	MountRequirements  types.Set     `tfsdk:"mount_requirements"`
	Name               types.String  `tfsdk:"name"`
	PortRequirements   types.Set     `tfsdk:"port_requirements"`
	Readme             types.String  `tfsdk:"readme"`
	ReceiveMessageType types.String  `tfsdk:"receive_message_type"`
	SendMessageType    types.String  `tfsdk:"send_message_type"`
}

type mountRequirementsModel struct {
	Description types.String `tfsdk:"description"`
	Source      types.String `tfsdk:"source"`
	Target      types.String `tfsdk:"target"`
}

type portRequirementsModel struct {
	ContainerPort types.Int64  `tfsdk:"container_port"`
	Description   types.String `tfsdk:"description"`
	Protocol      types.String `tfsdk:"protocol"`
}

func mountRequirementsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"description": types.StringType,
		"source":      types.StringType,
		"target":      types.StringType,
	}
}

func mountRequirementsAttrValues(
	description string,
	source *string,
	target string,
) map[string]attr.Value {
	s := types.StringNull()
	if source != nil {
		s = types.StringValue(*source)
	}
	return map[string]attr.Value{
		"description": types.StringValue(description),
		"source":      s,
		"target":      types.StringValue(target),
	}
}

func portRequirementAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"container_port": types.Int64Type,
		"description":    types.StringType,
		"protocol":       types.StringType,
	}
}

func portRequirementAttrValues(
	contanerPort int,
	description string,
	protocol api.Protocol,
) map[string]attr.Value {
	return map[string]attr.Value{
		"container_port": types.Int64Value(int64(contanerPort)),
		"description":    types.StringValue(description),
		"protocol":       types.StringValue(string(protocol)),
	}
}

func readManagedNodeType(ctx context.Context, client graphql.Client, name string, tenant string) (*managedNodeTypeModel, bool, diag.Diagnostics) {
	var (
		data   *managedNodeTypeModel
		diags  diag.Diagnostics
		system bool = false
	)

	if echoResp, err := api.ReadManagedNodeType(ctx, client, name, tenant); err == nil {
		if echoResp.GetManagedNodeType != nil {
			data = &managedNodeTypeModel{}
			if echoResp.GetManagedNodeType.ConfigTemplate != nil {
				data.ConfigTemplate = common.ConfigValue(*echoResp.GetManagedNodeType.ConfigTemplate)
			} else {
				data.ConfigTemplate = common.ConfigNull()
			}
			data.Description = types.StringValue(echoResp.GetManagedNodeType.Description)
			data.Id = types.StringValue(echoResp.GetManagedNodeType.Name)
			data.ImageUri = types.StringValue(echoResp.GetManagedNodeType.ImageUri)
			data.InUse = types.BoolValue(echoResp.GetManagedNodeType.InUse)
			if len(echoResp.GetManagedNodeType.MountRequirements) > 0 {
				elems := []attr.Value{}
				for _, mountReq := range echoResp.GetManagedNodeType.MountRequirements {
					if elem, d := types.ObjectValue(mountRequirementsAttrTypes(), mountRequirementsAttrValues(mountReq.Description, mountReq.Source, mountReq.Target)); d != nil {
						diags.Append(d...)
					} else {
						elems = append(elems, elem)
					}
				}
				var d diag.Diagnostics
				data.MountRequirements, d = types.SetValue(types.ObjectType{AttrTypes: mountRequirementsAttrTypes()}, elems)
				if d != nil && d.HasError() {
					diags.Append(d...)
				}
			} else {
				data.MountRequirements = types.SetNull(types.ObjectType{AttrTypes: mountRequirementsAttrTypes()})
			}
			data.Name = types.StringValue(echoResp.GetManagedNodeType.Name)
			if len(echoResp.GetManagedNodeType.PortRequirements) > 0 {
				elems := []attr.Value{}
				for _, portReq := range echoResp.GetManagedNodeType.PortRequirements {
					if elem, d := types.ObjectValue(portRequirementAttrTypes(), portRequirementAttrValues(portReq.ContainerPort, portReq.Description, portReq.Protocol)); d != nil {
						diags.Append(d...)
					} else {
						elems = append(elems, elem)
					}
				}
				var d diag.Diagnostics
				data.PortRequirements, d = types.SetValue(types.ObjectType{AttrTypes: portRequirementAttrTypes()}, elems)
				if d != nil && d.HasError() {
					diags.Append(d...)
				}
			} else {
				data.PortRequirements = types.SetNull(types.ObjectType{AttrTypes: portRequirementAttrTypes()})
			}
			if echoResp.GetManagedNodeType.Readme != nil {
				data.Readme = types.StringValue(*echoResp.GetManagedNodeType.Readme)
			} else {
				data.Readme = types.StringNull()
			}
			if echoResp.GetManagedNodeType.ReceiveMessageType != nil {
				data.ReceiveMessageType = types.StringValue(echoResp.GetManagedNodeType.ReceiveMessageType.Name)
			} else {
				data.ReceiveMessageType = types.StringNull()
			}
			if echoResp.GetManagedNodeType.SendMessageType != nil {
				data.SendMessageType = types.StringValue(echoResp.GetManagedNodeType.SendMessageType.Name)
			} else {
				data.SendMessageType = types.StringNull()
			}
			if echoResp.GetManagedNodeType.System != nil {
				system = *echoResp.GetManagedNodeType.System
			}
		}
	} else {
		diags.AddError("Error reading ManagedNodeType", err.Error())
	}

	return data, system, diags
}
