package tenant

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

type tenantModel struct {
	Active                 types.Bool    `tfsdk:"active"`
	AwsCredentials         types.Object  `tfsdk:"aws_credentials"`
	AwsCredentialsDuration types.Int64   `tfsdk:"aws_credentials_duration"`
	Config                 common.Config `tfsdk:"config"`
	Description            types.String  `tfsdk:"description"`
	Id                     types.String  `tfsdk:"id"`
	Name                   types.String  `tfsdk:"name"`
	Region                 types.String  `tfsdk:"region"`
	Table                  types.String  `tfsdk:"table"`
}

func tenantDataSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"active": {
			Computed:            true,
			MarkdownDescription: "The current Tenant's active state.",
			Type:                types.BoolType,
		},
		"aws_credentials": {
			Attributes:          tfsdk.SingleNestedAttributes(common.AwsCredentialsSchema()),
			Computed:            true,
			MarkdownDescription: "The AWS Session Credentials that allow the current ApiUser (configured in the provider) to access the Tenant's resources.",
		},
		"aws_credentials_duration": {
			MarkdownDescription: "The duration to request for `aws_credentials`. Must be set to obtain `aws_credentials`.",
			Optional:            true,
			Type:                types.Int64Type,
		},
		"config": {
			Computed:            true,
			MarkdownDescription: "The config for the Tenant. All nodes in the Tenant will be allowed to access this. Must be a JSON object.",
			Sensitive:           true,
			Type:                common.ConfigType{},
		},
		"description": {
			Computed:            true,
			MarkdownDescription: "A human-readable description.",
			Type:                types.StringType,
		},
		"id": {
			Computed: true,
			Type:     types.StringType,
		},
		"name": {
			Computed:            true,
			MarkdownDescription: "The name.",
			Type:                types.StringType,
		},
		"region": {
			Computed:            true,
			MarkdownDescription: "The current Tenant's AWS region name (e.g.  - `us-east-1`).",
			Type:                types.StringType,
		},
		"table": {
			Computed:            true,
			MarkdownDescription: "The current Tenant's DynamoDB [table](https://docs.echo.stream/docs/table) name.",
			Type:                types.StringType,
		},
	}
}

func tenantResourceSchema() map[string]tfsdk.Attribute {
	schema := tenantDataSchema()
	for key, attribute := range schema {
		if slices.Contains([]string{"config", "description"}, key) {
			attribute.Computed = false
			attribute.Optional = true
			schema[key] = attribute
		}
	}
	return schema
}

func readTenantData(ctx context.Context, client graphql.Client, tenant string, data *tenantModel) diag.Diagnostics {
	var (
		diags    diag.Diagnostics
		echoResp *api.ReadTenantResponse
		err      error
	)

	if echoResp, err = api.ReadTenant(ctx, client, tenant); err != nil {
		diags.AddError("Error reading Tenant data", err.Error())
	} else if echoResp.GetTenant == nil {
		diags.AddError("Tenant not found", fmt.Sprintf("Unable to find Tenant '%s'", tenant))
	} else {
		data.Active = types.Bool{Value: echoResp.GetTenant.Active}
		if echoResp.GetTenant.Config != nil {
			data.Config = common.Config{Value: *echoResp.GetTenant.Config}
		} else {
			data.Config = common.Config{Null: true}
		}
		if echoResp.GetTenant.Description != nil {
			data.Description = types.String{Value: *echoResp.GetTenant.Description}
		} else {
			data.Description = types.String{Null: true}
		}
		data.Id = types.String{Value: echoResp.GetTenant.Name}
		data.Name = types.String{Value: echoResp.GetTenant.Name}
		data.Region = types.String{Value: echoResp.GetTenant.Region}
		data.Table = types.String{Value: echoResp.GetTenant.Table}
		diags.Append(readTenantAwsCredentials(ctx, client, tenant, data)...)
	}

	return diags
}

func readTenantAwsCredentials(ctx context.Context, client graphql.Client, tenant string, data *tenantModel) diag.Diagnostics {
	var (
		diags    diag.Diagnostics
		echoResp *api.ReadTenantAwsCredentialsResponse
		err      error
	)

	if !(data.AwsCredentialsDuration.IsNull() || data.AwsCredentialsDuration.IsUnknown()) {
		duration := int(data.AwsCredentialsDuration.Value)
		if echoResp, err = api.ReadTenantAwsCredentials(ctx, client, tenant, &duration); err == nil {
			data.AwsCredentials = types.Object{
				Attrs: common.AwsCredentialsAttrValues(
					echoResp.GetTenant.GetAwsCredentials.AccessKeyId,
					echoResp.GetTenant.GetAwsCredentials.Expiration,
					echoResp.GetTenant.GetAwsCredentials.SecretAccessKey,
					echoResp.GetTenant.GetAwsCredentials.SessionToken,
				),
				AttrTypes: common.AwsCredentialsAttrTypes(),
			}
		}
	} else {
		data.AwsCredentials = types.Object{
			AttrTypes: common.AwsCredentialsAttrTypes(),
			Null:      true,
		}
	}

	if err != nil {
		diags.AddError("Error reading Tenant AWS Credentials", err.Error())
	}
	return diags
}
