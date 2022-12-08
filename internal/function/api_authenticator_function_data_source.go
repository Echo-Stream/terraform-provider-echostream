package function

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ datasource.DataSourceWithConfigure = &ApiAuthenticatorFunctionDataSource{}
	_ datasource.DataSourceWithSchema    = &ApiAuthenticatorFunctionDataSource{}
)

type ApiAuthenticatorFunctionDataSource struct {
	data *common.ProviderData
}

func (d *ApiAuthenticatorFunctionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApiAuthenticatorFunctionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_authenticator_function"
}

func (d *ApiAuthenticatorFunctionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config functionModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, _, diags := readApiAuthenicatorFunction(ctx, d.data.Client, config.Name.ValueString(), d.data.Tenant); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	} else if data == nil {
		resp.Diagnostics.AddError("ApiAuthenticatorFunction not found", fmt.Sprintf("Unable to find ApiAuthenticatorFunction '%s'", config.Name.ValueString()))
		return
	} else {
		config = *data
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *ApiAuthenticatorFunctionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:          dataFunctionAttributes(),
		MarkdownDescription: "[ApiAuthenticatorFunctions](https://docs.echo.stream/docs/webhook#api-authenticator-function) are managed Functions used in API-based Nodes (e.g. - WebhookNode).",
	}
}
