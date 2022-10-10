package function

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSourceWithConfigure = &BitmapperFunctionDataSource{}

type BitmapperFunctionDataSource struct {
	data *common.ProviderData
}

func (d *BitmapperFunctionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BitmapperFunctionDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:          dataBitmapperFunctionSchema(),
		Description:         "",
		MarkdownDescription: "",
	}, nil
}

func (d *BitmapperFunctionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bitmapper_function"
}

func (d *BitmapperFunctionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bitmapperFunctionModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, _, err := readBitmapperFunction(ctx, d.data.Client, config.Name.Value, d.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading Function", err.Error())
		return
	} else if data == nil {
		resp.Diagnostics.AddError("BitmapperFunction not found", fmt.Sprintf("Unable to find BitmapperFunction '%s'", config.Name.Value))
		return
	} else {
		config = *data
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
