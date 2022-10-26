package node

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigValidators = &WebhookNodeResource{}
	_ resource.ResourceWithImportState      = &WebhookNodeResource{}
)

// WebhookNodeResource defines the resource implementation.
type WebhookNodeResource struct {
	data *common.ProviderData
}

type webhookNodeModel struct {
	Config                  common.Config `tfsdk:"config"`
	Description             types.String  `tfsdk:"description"`
	Endpoint                types.String  `tfsdk:"endpoint"`
	InlineApiAuthenticator  types.String  `tfsdk:"inline_api_authenticator"`
	LoggingLevel            types.String  `tfsdk:"logging_level"`
	ManagedApiAuthenticator types.String  `tfsdk:"managed_api_authenticator"`
	Name                    types.String  `tfsdk:"name"`
	Requirements            types.Set     `tfsdk:"requirements"`
	SendMessageType         types.String  `tfsdk:"send_message_type"`
}

func (r *WebhookNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WebhookNodeResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("inline_api_authenticator"),
			path.MatchRoot("managed_api_authenticator"),
		),
	}
}

func (r *WebhookNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config                  *string
		description             *string
		inlineApiAuthenticator  *string
		loggingLevel            *api.LogLevel
		managedApiAuthenticator *string
		requirements            []string
		sendMessageType         *string
	)

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.InlineApiAuthenticator.IsNull() || plan.InlineApiAuthenticator.IsUnknown()) {
		inlineApiAuthenticator = &plan.InlineApiAuthenticator.Value
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		loggingLevel = (*api.LogLevel)(&plan.LoggingLevel.Value)
	}
	if !(plan.ManagedApiAuthenticator.IsNull() || plan.ManagedApiAuthenticator.IsUnknown()) {
		managedApiAuthenticator = &plan.ManagedApiAuthenticator.Value
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.SendMessageType.IsNull() || plan.SendMessageType.IsUnknown()) {
		sendMessageType = &plan.SendMessageType.Value
	}

	if echoResp, err := api.CreateWebhookNode(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		config,
		description,
		inlineApiAuthenticator,
		loggingLevel,
		managedApiAuthenticator,
		requirements,
		sendMessageType,
	); err != nil {
		resp.Diagnostics.AddError("Error creating WebhookNode", err.Error())
		return
	} else {
		if echoResp.CreateWebhookNode.Config != nil {
			plan.Config = common.Config{Value: *echoResp.CreateWebhookNode.Config}
		} else {
			plan.Config = common.Config{Null: true}
		}
		if echoResp.CreateWebhookNode.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateWebhookNode.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		plan.Endpoint = types.String{Value: echoResp.CreateWebhookNode.Endpoint}
		if echoResp.CreateWebhookNode.InlineApiAuthenticator != nil {
			plan.InlineApiAuthenticator = types.String{Value: *echoResp.CreateWebhookNode.InlineApiAuthenticator}
		} else {
			plan.InlineApiAuthenticator = types.String{Null: true}
		}
		if echoResp.CreateWebhookNode.LoggingLevel != nil {
			plan.LoggingLevel = types.String{Value: string(*echoResp.CreateWebhookNode.LoggingLevel)}
		} else {
			plan.LoggingLevel = types.String{Null: true}
		}
		if echoResp.CreateWebhookNode.ManagedApiAuthenticator != nil {
			plan.ManagedApiAuthenticator = types.String{Value: echoResp.CreateWebhookNode.ManagedApiAuthenticator.Name}
		} else {
			plan.ManagedApiAuthenticator = types.String{Null: true}
		}
		plan.Name = types.String{Value: echoResp.CreateWebhookNode.Name}
		plan.Requirements = types.Set{ElemType: types.StringType}
		if len(echoResp.CreateWebhookNode.Requirements) > 0 {
			for _, req := range echoResp.CreateWebhookNode.Requirements {
				plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
			}
		} else {
			plan.Requirements.Null = true
		}
		plan.SendMessageType = types.String{Value: echoResp.CreateWebhookNode.SendMessageType.Name}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WebhookNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting WebhookNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *WebhookNodeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := dataSendNodeSchema()
	for key, attribute := range schema {
		switch key {
		case "description":
			attribute.Computed = false
			attribute.Optional = true
		case "name":
			attribute.Computed = false
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
			attribute.Required = true
			attribute.Validators = common.FunctionNodeNameValidators
		case "send_message_type":
			attribute.Computed = true
			attribute.MarkdownDescription = "Must be JSON based, defaults to `echo.json`"
			attribute.Optional = true
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
		}
		schema[key] = attribute
	}
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"config": {
				MarkdownDescription: "The config, in JSON object format (i.e. - dict, map)",
				Optional:            true,
				Sensitive:           true,
				Type:                common.ConfigType{},
			},
			"endpoint": {
				Computed: true,
				MarkdownDescription: "The Webhooks endpoint to forward webhooks events to." +
					" Accepts POST webhook events at the root path." +
					" POST events may be any JSON-based payload",
				Type: types.StringType,
			},
			"inline_api_authenticator": {
				MarkdownDescription: "A Python code string that contains a single top-level function definition." +
					" This function must have the signature `(*, context, request, **kwargs)` and return" +
					" `None` or a tuple containing an `AuthCredentials` and `BaseUser` (or subclasses)." +
					" Mutually exclusive with `managedApiAuthenticator`",
				Optional: true,
				Type:     types.StringType,
			},
			"logging_level": {
				MarkdownDescription: "The logging level. One of `DEBUG`, `ERROR`, `INFO`, `WARNING`. Defaults to `INFO`",
				Optional:            true,
				Type:                types.StringType,
				Validators:          []tfsdk.AttributeValidator{common.LogLevelValidator},
			},
			"managed_api_authenticator": {
				MarkdownDescription: "The managedApiAuthenticator. Mutually exclusive with the `inlineApiAuthenticator`",
				Optional:            true,
				Type:                types.StringType,
			},
			"requirements": {
				MarkdownDescription: "The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format",
				Optional:            true,
				Type:                types.SetType{ElemType: types.StringType},
				Validators:          []tfsdk.AttributeValidator{common.RequirementsValidator},
			},
		},
	)
	return tfsdk.Schema{
		Attributes: schema,
		MarkdownDescription: "[WebhookNodes](https://docs.echo.stream/docs/webhook) allow for almost any processing " +
			"of messages, including transformation, augmentation, generation, combination and splitting",
	}, nil
}

func (r *WebhookNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *WebhookNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_node"
}

func (r *WebhookNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhookNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading WebhookNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeWebhookNode:
			if node.Config != nil {
				state.Config = common.Config{Value: *node.Config}
			} else {
				state.Config = common.Config{Null: true}
			}
			if node.Description != nil {
				state.Description = types.String{Value: *node.Description}
			} else {
				state.Description = types.String{Null: true}
			}
			state.Endpoint = types.String{Value: node.Endpoint}
			if node.InlineApiAuthenticator != nil {
				state.InlineApiAuthenticator = types.String{Value: *node.InlineApiAuthenticator}
			} else {
				state.InlineApiAuthenticator = types.String{Null: true}
			}
			if node.LoggingLevel != nil {
				state.LoggingLevel = types.String{Value: string(*node.LoggingLevel)}
			} else {
				state.LoggingLevel = types.String{Null: true}
			}
			if node.ManagedApiAuthenticator != nil {
				state.ManagedApiAuthenticator = types.String{Value: node.ManagedApiAuthenticator.Name}
			} else {
				state.ManagedApiAuthenticator = types.String{Null: true}
			}
			state.Name = types.String{Value: node.Name}
			state.Requirements = types.Set{ElemType: types.StringType}
			if len(node.Requirements) > 0 {
				for _, req := range node.Requirements {
					state.Requirements.Elems = append(state.Requirements.Elems, types.String{Value: req})
				}
			} else {
				state.Requirements.Null = true
			}
			state.SendMessageType = types.String{Value: node.SendMessageType.Name}
		default:
			resp.Diagnostics.AddError(
				"Expected WebhookNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WebhookNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config                  *string
		description             *string
		inlineApiAuthenticator  *string
		loggingLevel            *api.LogLevel
		managedApiAuthenticator *string
		requirements            []string
	)
	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.InlineApiAuthenticator.IsNull() || plan.InlineApiAuthenticator.IsUnknown()) {
		inlineApiAuthenticator = &plan.InlineApiAuthenticator.Value
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		loggingLevel = (*api.LogLevel)(&plan.LoggingLevel.Value)
	}
	if !(plan.ManagedApiAuthenticator.IsNull() || plan.ManagedApiAuthenticator.IsUnknown()) {
		managedApiAuthenticator = &plan.ManagedApiAuthenticator.Value
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	if echoResp, err := api.UpdateWebhookNode(
		ctx,
		r.data.Client,
		plan.Name.Value,
		r.data.Tenant,
		config,
		description,
		inlineApiAuthenticator,
		loggingLevel,
		managedApiAuthenticator,
		requirements,
	); err != nil {
		resp.Diagnostics.AddError("Error updating WebhookNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find WebhookNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.Value))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateWebhookNodeGetNodeWebhookNode:
			if node.Update.Config != nil {
				plan.Config = common.Config{Value: *node.Update.Config}
			} else {
				plan.Config = common.Config{Null: true}
			}
			if node.Update.Description != nil {
				plan.Description = types.String{Value: *node.Update.Description}
			} else {
				plan.Description = types.String{Null: true}
			}
			plan.Endpoint = types.String{Value: node.Update.Endpoint}
			if node.Update.InlineApiAuthenticator != nil {
				plan.InlineApiAuthenticator = types.String{Value: *node.Update.InlineApiAuthenticator}
			} else {
				plan.InlineApiAuthenticator = types.String{Null: true}
			}
			if node.Update.LoggingLevel != nil {
				plan.LoggingLevel = types.String{Value: string(*node.Update.LoggingLevel)}
			} else {
				plan.LoggingLevel = types.String{Null: true}
			}
			if node.Update.ManagedApiAuthenticator != nil {
				plan.ManagedApiAuthenticator = types.String{Value: node.Update.ManagedApiAuthenticator.Name}
			} else {
				plan.ManagedApiAuthenticator = types.String{Null: true}
			}
			plan.Name = types.String{Value: node.Update.Name}
			plan.Requirements = types.Set{ElemType: types.StringType}
			if len(node.Update.Requirements) > 0 {
				for _, req := range node.Update.Requirements {
					plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
				}
			} else {
				plan.Requirements.Null = true
			}
			plan.SendMessageType = types.String{Value: node.Update.SendMessageType.Name}
		default:
			resp.Diagnostics.AddError(
				"Expected WebhookNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
