package edge

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type edgeModel struct {
	Arn             types.String `tfsdk:"arn"`
	Description     types.String `tfsdk:"description"`
	KmsKey          types.String `tfsdk:"kmskey"`
	MaxReceiveCount types.Int64  `tfsdk:"max_receive_count"`
	MessageType     types.String `tfsdk:"message_type"`
	Queue           types.String `tfsdk:"queue"`
	Source          types.String `tfsdk:"source"`
	Target          types.String `tfsdk:"target"`
}
