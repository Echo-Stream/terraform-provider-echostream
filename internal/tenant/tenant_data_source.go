package tenant

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSourceWithConfigure = &TenantDataSource{}

type TenantDataSource struct {
	data *common.ProviderData
}

func (d *TenantDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TenantDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (d *TenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tenantModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(readTenantData(ctx, d.data.Client, d.data.Tenant, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *TenantDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"active": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "The current Tenant's active state.",
			},
			"aws_credentials": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"access_key_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The AWS Acces Key Id for the session.",
					},
					"expiration": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The date/time that the sesssion expires, in [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) format.",
					},
					"secret_access_key": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The AWS Secret Access Key for the session.",
						Sensitive:           true,
					},
					"session_token": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The AWS Session Token for the session.",
					},
				},
				Computed:            true,
				MarkdownDescription: "The AWS Session Credentials that allow the current ApiUser (configured in the provider) to access the Tenant's resources.",
			},
			"aws_credentials_duration": schema.Int64Attribute{
				MarkdownDescription: "The duration to request for `aws_credentials`. Must be set to obtain `aws_credentials`.",
				Optional:            true,
			},
			"config": schema.StringAttribute{
				Computed:            true,
				CustomType:          common.ConfigType{},
				MarkdownDescription: "The config for the Tenant. All nodes in the Tenant will be allowed to access this. Must be a JSON object.",
				Sensitive:           true,
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A human-readable description.",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name.",
			},
			"region": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current Tenant's AWS region name (e.g.  - `us-east-1`).",
			},
			"table": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current Tenant's DynamoDB [table](https://docs.echo.stream/docs/table) name.",
			},
		},
		MarkdownDescription: "Gets the current Tenant's information.",
	}
}
