package edge

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSourceWithConfigure = &EdgeDataSource{}

type EdgeDataSource struct {
	data *common.ProviderData
}

func (d *EdgeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EdgeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge"
}

func (d *EdgeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config edgeModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadEdge(ctx, d.data.Client, config.Source.ValueString(), config.Target.ValueString(), d.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading Edge", err.Error())
		return
	} else if echoResp.GetEdge == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		config.Arn = types.StringValue(echoResp.GetEdge.Arn)
		if echoResp.GetEdge.Description != nil {
			config.Description = types.StringValue(*echoResp.GetEdge.Description)
		} else {
			config.Description = types.StringNull()
		}
		if echoResp.GetEdge.KmsKey != nil {
			config.KmsKey = types.StringValue(echoResp.GetEdge.KmsKey.Name)
		} else {
			config.KmsKey = types.StringNull()
		}
		if echoResp.GetEdge.MaxReceiveCount != nil {
			config.MaxReceiveCount = types.Int64Value(int64(*echoResp.GetEdge.MaxReceiveCount))
		}
		config.MessageType = types.StringValue(echoResp.GetEdge.MessageType.Name)
		config.Queue = types.StringValue(echoResp.GetEdge.Queue)
		config.Source = types.StringValue(echoResp.GetEdge.Source.GetName())
		config.Target = types.StringValue(echoResp.GetEdge.Target.GetName())
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *EdgeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ARN of the underlying AWS SQS Queue.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A human-readable description.",
			},
			"kmskey": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the KmsKey to use to encrypt the message at rest and in flight. Defaults to the Tenant's KmsKey.",
			},
			"max_receive_count": schema.Int64Attribute{
				Computed: true,
				MarkdownDescription: "The maximum number of delivbery tries to the `target`. `0` is the default and will try forever. " +
					"Any positive number will result in that many tries before sending the messagge to the `DeadLetterEmitterNode`.",
			},
			"message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that will be transmitted.",
			},
			"queue": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL of the underlying AWS SQS queue.",
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "The source Node to transmit messages from.",
				Required:            true,
			},
			"target": schema.StringAttribute{
				MarkdownDescription: "The target Node to transmit messages to.",
				Required:            true,
			},
		},
		MarkdownDescription: "[Edges](https://docs.echo.stream/docs/edges) transmit messages of a single MessageType between Nodes.",
	}
}
