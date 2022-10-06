package function

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	echoResp, err := api.ReadFunction(ctx, d.data.Client, config.Name.Value, d.data.Tenant)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading Function %s", config.Name.Value), err.Error())
		return
	}
	if echoResp.GetFunction != nil {
		switch function := (*echoResp.GetFunction).(type) {
		case *api.ReadFunctionGetFunctionBitmapperFunction:
			config.ArgumentMessageType = types.String{Value: function.ArgumentMessageType.Name}
			config.Code = types.String{Value: function.Code}
			config.Description = types.String{Value: function.Description}
			config.InUse = types.Bool{Value: function.InUse}
			config.Name = types.String{Value: function.Name}
			if function.Readme != nil {
				config.Readme = types.String{Value: *function.Readme}
			} else {
				config.Readme = types.String{Null: true}
			}
			if len(function.Requirements) > 0 {
				config.Requirements = types.Set{ElemType: types.StringType}
				for _, req := range function.Requirements {
					config.Requirements.Elems = append(config.Requirements.Elems, types.String{Value: req})
				}
			} else {
				config.Requirements.Null = true
			}
		default:
			resp.Diagnostics.AddError(
				"Incorrect Function type",
				fmt.Sprintf("expected BitmapperFunction, got %v", function.GetTypename()),
			)
			return
		}
	} else {
		resp.Diagnostics.AddWarning("Unable to find function", fmt.Sprintf("'%s' function does not exist", config.Name.Value))
		config = bitmapperFunctionModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
