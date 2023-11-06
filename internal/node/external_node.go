package node

import (
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type externalNodeModel struct {
	App                types.String  `tfsdk:"app"`
	Config             common.Config `tfsdk:"config"`
	Description        types.String  `tfsdk:"description"`
	Name               types.String  `tfsdk:"name"`
	ReceiveMessageType types.String  `tfsdk:"receive_message_type"`
	SendMessageType    types.String  `tfsdk:"send_message_type"`
}
