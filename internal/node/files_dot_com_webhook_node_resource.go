package node

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithImportState = &FilesDotComWebhookNodeResource{}
)

// FilesDotComWebhookNodeResource defines the resource implementation.
type FilesDotComWebhookNodeResource struct {
	data *common.ProviderData
}

type filesDotComWebhookNodeModel struct {
	ApiKey          types.String `tfsdk:"api_key"`
	Description     types.String `tfsdk:"description"`
	Endpoint        types.String `tfsdk:"endpoint"`
	Name            types.String `tfsdk:"name"`
	SendMessageType types.String `tfsdk:"send_message_type"`
	Token           types.String `tfsdk:"token"`
}

func (r *FilesDotComWebhookNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*common.ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.data = data
}

func (r *FilesDotComWebhookNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan filesDotComWebhookNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}

	if echoResp, err := api.CreateFilesDotComWebhookNode(
		ctx,
		r.data.Client,
		plan.ApiKey.Value,
		plan.Name.Value,
		r.data.Tenant,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error creating FilesDotComWebhookNode", err.Error())
		return
	} else {
		if echoResp.CreateFilesDotComWebhookNode.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateFilesDotComWebhookNode.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		plan.Endpoint = types.String{Value: echoResp.CreateFilesDotComWebhookNode.Endpoint}
		plan.Name = types.String{Value: echoResp.CreateFilesDotComWebhookNode.Name}
		plan.SendMessageType = types.String{Value: echoResp.CreateFilesDotComWebhookNode.SendMessageType.Name}
		plan.Token = types.String{Value: echoResp.CreateFilesDotComWebhookNode.Token}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FilesDotComWebhookNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state filesDotComWebhookNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting FilesDotComWebhookNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *FilesDotComWebhookNodeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := dataSendNodeSchema()
	for key, attribute := range schema {
		switch key {
		case "description":
			attribute.Computed = false
			attribute.Optional = true
		case "name":
			attribute.Computed = false
			attribute.Required = true
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
			attribute.Validators = common.NameValidators
		}
		schema[key] = attribute
	}
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"api_key": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				Sensitive:           true,
				Type:                types.StringType,
				Validators:          []tfsdk.AttributeValidator{stringvalidator.LengthAtLeast(1)},
			},
			"endpoint": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
			"token": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Sensitive:           true,
				Type:                types.StringType,
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		Description:         "FilesDotComWebhookNodes receive webhooks from Files.com (https://www.files.com)",
		MarkdownDescription: "FilesDotComWebhookNodes receive webhooks from [Files.com](https://www.files.com)",
	}, nil
}

func (r *FilesDotComWebhookNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *FilesDotComWebhookNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_files_dot_com_webhook_node"
}

func (r *FilesDotComWebhookNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state filesDotComWebhookNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading FilesDotComWebhookNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeFilesDotComWebhookNode:
			if node.Description != nil {
				state.Description = types.String{Value: *node.Description}
			} else {
				state.Description = types.String{Null: true}
			}
			state.Endpoint = types.String{Value: node.Endpoint}
			state.Name = types.String{Value: node.Name}
			state.SendMessageType = types.String{Value: node.SendMessageType.Name}
			state.Token = types.String{Value: node.Token}
		default:
			resp.Diagnostics.AddError(
				"Expected FilesDotComWebhookNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FilesDotComWebhookNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan filesDotComWebhookNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var description *string
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}

	if echoResp, err := api.UpdateFilesDotComWebhookNode(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		&plan.ApiKey.Value,
		description,
	); err != nil {
		resp.Diagnostics.AddError("Error updating FilesDotComWebhookNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find FilesDotComWebhookNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.Value))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateFilesDotComWebhookNodeGetNodeFilesDotComWebhookNode:
			if node.Update.Description != nil {
				plan.Description = types.String{Value: *node.Update.Description}
			} else {
				plan.Description = types.String{Null: true}
			}
			plan.Endpoint = types.String{Value: node.Update.Endpoint}
			plan.Name = types.String{Value: node.Update.Name}
			plan.SendMessageType = types.String{Value: node.Update.SendMessageType.Name}
			plan.Token = types.String{Value: node.Update.Token}
		default:
			resp.Diagnostics.AddError(
				"Expected FilesDotComWebhookNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
