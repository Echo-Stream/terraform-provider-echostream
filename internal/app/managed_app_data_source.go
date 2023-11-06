package app

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSourceWithConfigure = &ManagedAppDataSource{}

type ManagedAppDataSource struct {
	data *common.ProviderData
}

func (d *ManagedAppDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*common.ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.data = data
}

func (d *ManagedAppDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_app"
}

func (d *ManagedAppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config managedAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadApp(ctx, d.data.Client, config.Name.ValueString(), d.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading ManagedApp", err.Error())
		return
	} else if echoResp.GetApp != nil {
		switch app := (*echoResp.GetApp).(type) {
		case *api.ReadAppGetAppManagedApp:
			config.AuditRecordsEndpoint = types.StringValue(app.AuditRecordsEndpoint)
			config.Name = types.StringValue(app.Name)
			if app.Config != nil {
				config.Config = common.ConfigValue(*app.Config)
			} else {
				config.Config = common.ConfigNull()
			}
			var diags diag.Diagnostics
			config.Credentials, diags = types.ObjectValue(
				common.CognitoCredentialsAttrTypes(),
				common.CognitoCredentialsAttrValues(
					app.Credentials.ClientId,
					app.Credentials.Password,
					app.Credentials.UserPoolId,
					app.Credentials.Username,
				),
			)
			if diags != nil && diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
			if app.Description != nil {
				config.Description = types.StringValue(*app.Description)
			} else {
				config.Description = types.StringNull()
			}
			config.Name = types.StringValue(app.Name)
			config.TableAccess = types.BoolValue(app.TableAccess)
		default:
			resp.Diagnostics.AddError(
				"Incorrect App type",
				fmt.Sprintf("'%s' is incorrect App type", config.Name.String()),
			)
			return
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *ManagedAppDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:          remoteAppDataSourceAttributes(),
		MarkdownDescription: "[ManagedApps](https://docs.echo.stream/docs/managed-app) provide fully managed (by EchoStream) processing resources in a remote virtual compute environment that you manage.",
	}
}
