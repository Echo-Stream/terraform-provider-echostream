package node

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"golang.org/x/exp/maps"
)

func nodeDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A human-readable description.",
		},
		"name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The name of the Node. Must be unique within the Tenant.",
		},
	}
}

func receiveNodeDataSourceAttributes() map[string]schema.Attribute {
	attributes := nodeDataSourceAttributes()
	maps.Copy(
		attributes,
		map[string]schema.Attribute{
			"receive_message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that this Node is capable of receiving.",
			},
		},
	)
	return attributes
}

func sendNodeDataSourceAttributes() map[string]schema.Attribute {
	attributes := nodeDataSourceAttributes()
	maps.Copy(
		attributes,
		map[string]schema.Attribute{
			"send_message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that this Node is capable of sending.",
			},
		},
	)
	return attributes
}

func sendReceiveNodeDataSourceAttributes() map[string]schema.Attribute {
	attributes := receiveNodeDataSourceAttributes()
	maps.Copy(attributes, sendNodeDataSourceAttributes())
	return attributes
}
