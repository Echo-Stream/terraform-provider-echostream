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
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSourceWithConfigure = &CrossAccountAppDataSource{}

type CrossAccountAppDataSource struct {
	data *common.ProviderData
}

func (d *CrossAccountAppDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CrossAccountAppDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cross_account_app"
}

func (d *CrossAccountAppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config crossAccountAppModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadApp(ctx, d.data.Client, config.Name.ValueString(), d.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading CrossAccountApp", err.Error())
		return
	} else if echoResp.GetApp != nil {
		switch app := (*echoResp.GetApp).(type) {
		case *api.ReadAppGetAppCrossAccountApp:
			config.Account = types.StringValue(app.Account)
			config.AppsyncEndpoint = types.StringValue(app.AppsyncEndpoint)
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
			config.IamPolicy = types.StringValue(app.IamPolicy)
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

func (d *CrossAccountAppDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := remoteAppDataSourceAttributes()
	maps.Copy(
		attributes,
		map[string]schema.Attribute{
			"account": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The AWS account number that will host this CrossAcountApp's compute resources.",
			},
			"appsync_endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The EchoStream AppSync Endpoint that this ExternalApp must use.",
			},
			"iam_policy": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The IAM policy to apply to this CrossAccountApp's compute resources (e.g. - Lambda, EC2) to grant access to its EchoStream resources.",
			},
		},
	)
	resp.Schema = schema.Schema{
		Attributes:          attributes,
		MarkdownDescription: "[CrossAccountApps](https://docs.echo.stream/docs/cross-account-app) provides a way to receive/send messages in their Nodes using cross-account IAM access.",
	}
}
