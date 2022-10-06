package tenant

import (
	"context"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

type tenantModel struct {
	Active      types.Bool         `tfsdk:"active"`
	Config      common.ConfigValue `tfsdk:"config"`
	Description types.String       `tfsdk:"description"`
	Name        types.String       `tfsdk:"name"`
	Region      types.String       `tfsdk:"region"`
	Table       types.String       `tfsdk:"table"`
}

func tenantDataSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"active": {
			Computed:            true,
			Description:         "The current Tenant's active state",
			MarkdownDescription: "The current Tenant's active state",
			Type:                types.BoolType,
		},
		"config": {
			Computed:            true,
			Description:         "",
			MarkdownDescription: "",
			Sensitive:           true,
			Type:                common.ConfigType{},
		},
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
		"region": {
			Computed:            true,
			Description:         "The current Tenant's AWS region name",
			MarkdownDescription: "The current Tenant's AWS region name (e.g.  - `us-east-1`)",
			Type:                types.StringType,
		},
		"table": {
			Computed:            true,
			Description:         "The current Tenant's DynamoDB tabel name",
			MarkdownDescription: "The current Tenant's DynamoDB tabel name",
			Type:                types.StringType,
		},
	}
}

func tenantResourceSchema() map[string]tfsdk.Attribute {
	schema := tenantDataSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"config": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Sensitive:           true,
				Type:                common.ConfigType{},
			},
			"description": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
			},
		},
	)
	return schema
}

func readTenantData(ctx context.Context, client graphql.Client, tenant string, data *tenantModel) error {
	var (
		echoResp *api.ReadTenantResponse
		err      error
	)

	if echoResp, err = api.ReadTenant(ctx, client, tenant); err == nil {
		data.Active = types.Bool{Value: echoResp.GetTenant.Active}
		if echoResp.GetTenant.Description != nil {
			data.Description = types.String{Value: *echoResp.GetTenant.Description}
		} else {
			data.Description = types.String{Null: true}
		}
		if echoResp.GetTenant.Config != nil {
			data.Config = common.ConfigValue{Value: *echoResp.GetTenant.Config}
		} else {
			data.Config = common.ConfigValue{Null: true}
		}
		data.Name = types.String{Value: echoResp.GetTenant.Name}
		data.Region = types.String{Value: echoResp.GetTenant.Region}
		data.Table = types.String{Value: echoResp.GetTenant.Table}
	}

	return err
}
