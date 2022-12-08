package tenant

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		data.Active = types.BoolValue(echoResp.GetTenant.Active)
		if echoResp.GetTenant.Config != nil {
			data.Config = common.ConfigValue(*echoResp.GetTenant.Config)
		} else {
			data.Config = common.ConfigNull()
		}
		if echoResp.GetTenant.Description != nil {
			data.Description = types.StringValue(*echoResp.GetTenant.Description)
		} else {
			data.Description = types.StringNull()
		}
		data.Id = types.StringValue(echoResp.GetTenant.Name)
		data.Name = types.StringValue(echoResp.GetTenant.Name)
		data.Region = types.StringValue(echoResp.GetTenant.Region)
		data.Table = types.StringValue(echoResp.GetTenant.Table)
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
		duration := int(data.AwsCredentialsDuration.ValueInt64())
		if echoResp, err = api.ReadTenantAwsCredentials(ctx, client, tenant, &duration); err == nil {
			var d diag.Diagnostics
			data.AwsCredentials, d = types.ObjectValue(
				common.AwsCredentialsAttrTypes(),
				common.AwsCredentialsAttrValues(
					echoResp.GetTenant.GetAwsCredentials.AccessKeyId,
					echoResp.GetTenant.GetAwsCredentials.Expiration,
					echoResp.GetTenant.GetAwsCredentials.SecretAccessKey,
					echoResp.GetTenant.GetAwsCredentials.SessionToken,
				),
			)
			if d != nil {
				diags = d
			}

		}
	} else {
		data.AwsCredentials = types.ObjectNull(common.AwsCredentialsAttrTypes())
	}

	if err != nil {
		diags.AddError("Error reading Tenant AWS Credentials", err.Error())
	}
	return diags
}
