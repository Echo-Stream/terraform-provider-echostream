package app

import (
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type managedAppModel struct {
	AuditRecordsEndpoint types.String  `tfsdk:"audit_records_endpoint"`
	Config               common.Config `tfsdk:"config"`
	Credentials          types.Object  `tfsdk:"credentials"`
	Description          types.String  `tfsdk:"description"`
	Name                 types.String  `tfsdk:"name"`
	TableAccess          types.Bool    `tfsdk:"table_access"`
}
