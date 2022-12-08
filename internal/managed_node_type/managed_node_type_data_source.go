package managed_node_type

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ datasource.DataSourceWithConfigure = &ManagedNodeTypeDataSource{}
	_ datasource.DataSourceWithSchema    = &ManagedNodeTypeDataSource{}
)

type ManagedNodeTypeDataSource struct {
	data *common.ProviderData
}

func (d *ManagedNodeTypeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ManagedNodeTypeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_node_type"
}

func (d *ManagedNodeTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config managedNodeTypeModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, _, diags := readManagedNodeType(ctx, d.data.Client, config.Name.ValueString(), d.data.Tenant); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	} else if data == nil {
		resp.Diagnostics.AddError("ManagedNodeType not found", fmt.Sprintf("Unable to find ManagedNodeType '%s'", config.Name.ValueString()))
		return
	} else {
		config = *data
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *ManagedNodeTypeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config_template": schema.StringAttribute{
				Computed:   true,
				CustomType: common.ConfigType{},
				MarkdownDescription: "A [JSON Schema](https://json-schema.org/) document that specifies the" +
					" requirements for the config attribute of ManagedNodes created using this ManagedNodeType.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A human-readable description.",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"image_uri": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The URI of the Docker image. Must be a [public](https://docs.aws.amazon.com/AmazonECR/latest/public/public-repositories.html) " +
					"or a [private](https://docs.aws.amazon.com/AmazonECR/latest/userguide/Repositories.html) AWS ECR repository.",
			},
			"in_use": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: " True if this is used by ManagedNodes.",
			},
			"mount_requirements": schema.SetNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The mount (i.e. - volume) requirements of the Docker image.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A human-readable description of the port.",
						},
						"source": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The path of the mount on the host.",
						},
						"target": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The path of the mount in the Docker container.",
						},
					},
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the ManagedNodeType. Must be unique within the Tenant.",
				Required:            true,
				Validators:          common.NameValidators,
			},
			"port_requirements": schema.SetNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The port requirements of the Docker image.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"container_port": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The exposed container port.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A human-readable description for the port.",
						},
						"protocol": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The protocol to use for the port. One of `sctp`, `tcp` or `udp`.",
						},
					},
				},
			},
			"readme": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "README in MarkDown format.",
			},
			"receive_message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that ManagedNodes created with this ManagedNodeType are capable of receiving.",
			},
			"send_message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that ManagedNodes created with this ManagedNodeType are capable of sending.",
			},
		},
		MarkdownDescription: "ManagedNodeTypes are wrappers around Docker image definitions and define the requirements " +
			"necessary to instantiate those images as Docker containers inside a ManagedNode.",
	}
}
