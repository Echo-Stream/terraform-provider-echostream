package node

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

func dataNodeSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"description": {
			Computed:            true,
			Description:         "",
			MarkdownDescription: "",
			Type:                types.StringType,
		},
		"name": {
			Computed:            true,
			Description:         "",
			MarkdownDescription: "",
			Type:                types.StringType,
		},
	}
}

func dataReceiveNodeSchema() map[string]tfsdk.Attribute {
	schema := dataNodeSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"receive_message_type": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
		},
	)
	return schema
}

func dataSendNodeSchema() map[string]tfsdk.Attribute {
	schema := dataNodeSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"send_message_type": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
		},
	)
	return schema
}

func dataSendReceiveNodeSchema() map[string]tfsdk.Attribute {
	schema := dataReceiveNodeSchema()
	maps.Copy(schema, dataSendNodeSchema())
	return schema
}
