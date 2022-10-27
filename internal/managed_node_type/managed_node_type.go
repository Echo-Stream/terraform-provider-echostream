package managed_node_type

import (
	"context"
	"regexp"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

type managedNodeTypeModel struct {
	ConfigTemplate     common.Config `tfsdk:"config_template"`
	Description        types.String  `tfsdk:"description"`
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

func dataManagedNodeTypeSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"config_template": {
			Computed: true,
			MarkdownDescription: "A [JSON Schema](https://json-schema.org/) document that specifies the" +
				" requirements for the config attribute of ManagedNodes created using this ManagedNodeType.",
			Type: common.ConfigType{},
		},
		"description": {
			Computed:            true,
			MarkdownDescription: "A human-readable description.",
			Type:                types.StringType,
		},
		"image_uri": {
			Computed: true,
			MarkdownDescription: "The URI of the Docker image. Must be a [public](https://docs.aws.amazon.com/AmazonECR/latest/public/public-repositories.html) " +
				"or a [private](https://docs.aws.amazon.com/AmazonECR/latest/userguide/Repositories.html) AWS ECR repository.",
			Type: types.StringType,
		},
		"in_use": {
			Computed:            true,
			MarkdownDescription: " True if this is used by ManagedNodes.",
			Type:                types.BoolType,
		},
		"mount_requirements": {
			Attributes:          tfsdk.SetNestedAttributes(dataMountRequirementsSchema()),
			Computed:            true,
			MarkdownDescription: "The mount (i.e. - volume) requirements of the Docker image.",
		},
		"name": {
			MarkdownDescription: "The name of the ManagedNodeType. Must be unique within the Tenant.",
			Required:            true,
			Type:                types.StringType,
			Validators:          common.NameValidators,
		},
		"port_requirements": {
			Attributes:          tfsdk.SetNestedAttributes(dataPortRequirementsSchema()),
			Computed:            true,
			MarkdownDescription: "The port requirements of the Docker image.",
		},
		"readme": {
			Computed:            true,
			MarkdownDescription: "README in MarkDown format.",
			Type:                types.StringType,
		},
		"receive_message_type": {
			Computed:            true,
			MarkdownDescription: "The MessageType that ManagedNodes created with this ManagedNodeType are capable of receiving.",
			Type:                types.StringType,
		},
		"send_message_type": {
			Computed:            true,
			MarkdownDescription: "The MessageType that ManagedNodes created with this ManagedNodeType are capable of sending.",
			Type:                types.StringType,
		},
	}
}

func dataMountRequirementsSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"description": {
			Computed:            true,
			MarkdownDescription: "A human-readable description of the port.",
			Type:                types.StringType,
		},
		"source": {
			Computed:            true,
			MarkdownDescription: "The path of the mount on the host.",
			Type:                types.StringType,
		},
		"target": {
			Computed:            true,
			MarkdownDescription: "The path of the mount in the Docker container.",
			Type:                types.StringType,
		},
	}
}

func dataPortRequirementsSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"container_port": {
			Computed:            true,
			MarkdownDescription: "The exposed container port.",
			Type:                types.Int64Type,
		},
		"description": {
			Computed:            true,
			MarkdownDescription: "A human-readable description for the port.",
			Type:                types.StringType,
		},
		"protocol": {
			Computed:            true,
			MarkdownDescription: "The protocol to use for the port. One of `sctp`, `tcp` or `udp`.",
			Type:                types.StringType,
		},
	}
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
	s := types.String{}
	if source != nil {
		s.Value = *source
	} else {
		s.Null = true
	}
	return map[string]attr.Value{
		"description": types.String{Value: description},
		"source":      s,
		"target":      types.String{Value: target},
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
		"container_port": types.Int64{Value: int64(contanerPort)},
		"description":    types.String{Value: description},
		"protocol":       types.String{Value: string(protocol)},
	}
}

func resourceManagedNodeTypeSchema() map[string]tfsdk.Attribute {
	schema := dataManagedNodeTypeSchema()
	for key, attribute := range schema {
		if key != "in_use" {
			attribute.Computed = false
			if slices.Contains([]string{"description", "image_uri", "name"}, key) {
				attribute.Required = true
			} else {
				attribute.Optional = true
			}
			if slices.Contains(
				[]string{
					"config_template",
					"mount_requirements",
					"name",
					"port_requirements",
					"receive_message_type",
					"send_message_type",
				}, key) {
				attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
			}
			switch key {
			case "image_uri":
				attribute.Validators = []tfsdk.AttributeValidator{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(?:(?:[0-9]{12}\.dkr\.ecr\.[a-z]+\-[a-z]+\-[0-9]\.amazonaws\.com/.+\:.+)|(?:public\.ecr\.aws/.+/.+\:.+))$`),
						`must be either a private ECR image URI (aws_account_id.dkr.ecr.region.amazonaws.com/respository:tag) or a public ECR image URI (public.ecr.aws/registry_alias/repository:tag)`,
					),
				}
			case "mount_requirements":
				attribute.Attributes = tfsdk.SetNestedAttributes(resourceMountRequirementsSchema())
			case "name":
				attribute.Validators = append(common.NameValidators, common.NotSystemNameValidator)
			case "port_requirements":
				attribute.Attributes = tfsdk.SetNestedAttributes(resourcePortRequirementsSchema())
			}
		}
		schema[key] = attribute
	}
	return schema
}

func resourceMountRequirementsSchema() map[string]tfsdk.Attribute {
	schema := dataMountRequirementsSchema()
	for key, attribute := range schema {
		attribute.Computed = false
		switch key {
		case "description", "target":
			attribute.Required = true
		case "source":
			attribute.Optional = true
		}
		schema[key] = attribute
	}
	return schema
}

func resourcePortRequirementsSchema() map[string]tfsdk.Attribute {
	schema := dataPortRequirementsSchema()
	for key, attribute := range schema {
		attribute.Computed = false
		attribute.Required = true
		if key == "protocol" {
			attribute.Validators = []tfsdk.AttributeValidator{common.ProtocolValidator}
		}
		schema[key] = attribute
	}
	return schema
}

func readManagedNodeType(ctx context.Context, client graphql.Client, name string, tenant string) (*managedNodeTypeModel, bool, error) {
	var (
		data     *managedNodeTypeModel
		echoResp *api.ReadManagedNodeTypeResponse
		err      error
		system   bool = false
	)

	if echoResp, err = api.ReadManagedNodeType(ctx, client, name, tenant); err == nil {
		if echoResp.GetManagedNodeType != nil {
			data = &managedNodeTypeModel{}
			if echoResp.GetManagedNodeType.ConfigTemplate != nil {
				data.ConfigTemplate = common.Config{Value: *echoResp.GetManagedNodeType.ConfigTemplate}
			} else {
				data.ConfigTemplate = common.Config{Null: true}
			}
			data.Description = types.String{Value: echoResp.GetManagedNodeType.Description}
			data.ImageUri = types.String{Value: echoResp.GetManagedNodeType.ImageUri}
			data.InUse = types.Bool{Value: echoResp.GetManagedNodeType.InUse}
			data.MountRequirements = types.Set{ElemType: types.ObjectType{AttrTypes: mountRequirementsAttrTypes()}}
			if len(echoResp.GetManagedNodeType.MountRequirements) > 0 {
				for _, mountReq := range echoResp.GetManagedNodeType.MountRequirements {
					data.MountRequirements.Elems = append(
						data.MountRequirements.Elems,
						types.Object{
							Attrs:     mountRequirementsAttrValues(mountReq.Description, mountReq.Source, mountReq.Target),
							AttrTypes: mountRequirementsAttrTypes(),
						},
					)
				}
			} else {
				data.MountRequirements.Null = true
			}
			data.Name = types.String{Value: echoResp.GetManagedNodeType.Name}
			data.PortRequirements = types.Set{ElemType: types.ObjectType{AttrTypes: portRequirementAttrTypes()}}
			if len(echoResp.GetManagedNodeType.PortRequirements) > 0 {
				for _, portReq := range echoResp.GetManagedNodeType.PortRequirements {
					data.PortRequirements.Elems = append(
						data.PortRequirements.Elems,
						types.Object{
							Attrs:     portRequirementAttrValues(portReq.ContainerPort, portReq.Description, portReq.Protocol),
							AttrTypes: portRequirementAttrTypes(),
						},
					)
				}
			} else {
				data.PortRequirements.Null = true
			}
			if echoResp.GetManagedNodeType.Readme != nil {
				data.Readme = types.String{Value: *echoResp.GetManagedNodeType.Readme}
			} else {
				data.Readme = types.String{Null: true}
			}
			if echoResp.GetManagedNodeType.ReceiveMessageType != nil {
				data.ReceiveMessageType = types.String{Value: echoResp.GetManagedNodeType.ReceiveMessageType.Name}
			} else {
				data.ReceiveMessageType = types.String{Null: true}
			}
			if echoResp.GetManagedNodeType.SendMessageType != nil {
				data.SendMessageType = types.String{Value: echoResp.GetManagedNodeType.SendMessageType.Name}
			} else {
				data.SendMessageType = types.String{Null: true}
			}
			if echoResp.GetManagedNodeType.System != nil {
				system = *echoResp.GetManagedNodeType.System
			}
		}
	}

	return data, system, err
}
