package node

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ConfigValidator              = &LeaseSecondsValidator{}
	_ resource.ResourceWithConfigure        = &WebSubHubNodeResource{}
	_ resource.ResourceWithConfigValidators = &WebSubHubNodeResource{}
	_ resource.ResourceWithImportState      = &WebSubHubNodeResource{}
)

type LeaseSecondsValidator struct{}

func (v LeaseSecondsValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v LeaseSecondsValidator) MarkdownDescription(_ context.Context) string {
	return "max_lease_seconds must be greater than or equal to default_lease_seconds"
}

func (v LeaseSecondsValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var (
		default_lease_seconds attr.Value
		max_lease_seconds     attr.Value
	)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("default_lease_seconds"), &default_lease_seconds)...)
	if default_lease_seconds.IsNull() || default_lease_seconds.IsUnknown() {
		return
	}
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("max_lease_seconds"), &max_lease_seconds)...)
	if !(max_lease_seconds.IsNull() || max_lease_seconds.IsUnknown()) {
		if default_lease_seconds.(types.Int64).ValueInt64() > max_lease_seconds.(types.Int64).ValueInt64() {
			resp.Diagnostics.AddAttributeError(
				path.Root("max_lease_seconds"),
				"max_lease_seconds must be greater than or equal to default_lease_seconds",
				"max_lease_seconds must be greater than or equal to default_lease_seconds",
			)
		}
	}
}

// WebSubHubNodeResource defines the resource implementation.
type WebSubHubNodeResource struct {
	data *common.ProviderData
}

type webSubHubNodeModel struct {
	Config                  common.Config `tfsdk:"config"`
	DefaultLeaseSeconds     types.Int64   `tfsdk:"default_lease_seconds"`
	DeliveryRetries         types.Int64   `tfsdk:"delivery_retries"`
	Description             types.String  `tfsdk:"description"`
	Endpoint                types.String  `tfsdk:"endpoint"`
	Id                      types.String  `tfsdk:"id"`
	InlineApiAuthenticator  types.String  `tfsdk:"inline_api_authenticator"`
	LoggingLevel            types.String  `tfsdk:"logging_level"`
	ManagedApiAuthenticator types.String  `tfsdk:"managed_api_authenticator"`
	MaxLeaseSeconds         types.Int64   `tfsdk:"max_lease_seconds"`
	Name                    types.String  `tfsdk:"name"`
	ReceiveMessageType      types.String  `tfsdk:"receive_message_type"`
	Requirements            types.Set     `tfsdk:"requirements"`
	SignatureAlgorithm      types.String  `tfsdk:"signature_algorithm"`
	SubscriptionSecurity    types.String  `tfsdk:"subscription_security"`
}

func (r *WebSubHubNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WebSubHubNodeResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("inline_api_authenticator"),
			path.MatchRoot("managed_api_authenticator"),
		),
		&LeaseSecondsValidator{},
	}
}

func (r *WebSubHubNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webSubHubNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config                  *string
		defaultLeaseSeconds     *int
		deliveryRetries         *int
		description             *string
		diags                   diag.Diagnostics
		inlineApiAuthenticator  *string
		loggingLevel            *api.LogLevel
		managedApiAuthenticator *string
		maxLeaseSeconds         *int
		requirements            []string
		signatureAlgorithm      *api.WebSubSignatureAlgorithm
		subscriptionSecurity    *api.WebSubSubscriptionSecurity
	)

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.DefaultLeaseSeconds.IsNull() || plan.DefaultLeaseSeconds.IsUnknown()) {
		temp := int(plan.DefaultLeaseSeconds.ValueInt64())
		defaultLeaseSeconds = &temp
	}
	if !(plan.DeliveryRetries.IsNull() || plan.DeliveryRetries.IsUnknown()) {
		temp := int(plan.DeliveryRetries.ValueInt64())
		deliveryRetries = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.InlineApiAuthenticator.IsNull() || plan.InlineApiAuthenticator.IsUnknown()) {
		temp := plan.InlineApiAuthenticator.ValueString()
		inlineApiAuthenticator = &temp
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		temp := plan.LoggingLevel.ValueString()
		loggingLevel = (*api.LogLevel)(&temp)
	}
	if !(plan.ManagedApiAuthenticator.IsNull() || plan.ManagedApiAuthenticator.IsUnknown()) {
		temp := plan.ManagedApiAuthenticator.ValueString()
		managedApiAuthenticator = &temp
	}
	if !(plan.MaxLeaseSeconds.IsNull() || plan.MaxLeaseSeconds.IsUnknown()) {
		temp := int(plan.MaxLeaseSeconds.ValueInt64())
		maxLeaseSeconds = &temp
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.SignatureAlgorithm.IsNull() || plan.SignatureAlgorithm.IsUnknown()) {
		temp := plan.SignatureAlgorithm.ValueString()
		signatureAlgorithm = (*api.WebSubSignatureAlgorithm)(&temp)
	}

	if echoResp, err := api.CreateWebSubHubNode(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		r.data.Tenant,
		config,
		defaultLeaseSeconds,
		deliveryRetries,
		description,
		inlineApiAuthenticator,
		loggingLevel,
		managedApiAuthenticator,
		maxLeaseSeconds,
		requirements,
		signatureAlgorithm,
		subscriptionSecurity,
	); err != nil {
		resp.Diagnostics.AddError("Error creating WebSubHubNode", err.Error())
		return
	} else {
		if echoResp.CreateWebSubHubNode.Config != nil {
			plan.Config = common.ConfigValue(*echoResp.CreateWebSubHubNode.Config)
		} else {
			plan.Config = common.ConfigNull()
		}
		plan.DefaultLeaseSeconds = types.Int64Value(int64(echoResp.CreateWebSubHubNode.DefaultLeaseSeconds))
		if echoResp.CreateWebSubHubNode.DeliveryRetries != nil {
			plan.DeliveryRetries = types.Int64Value(int64(*echoResp.CreateWebSubHubNode.DeliveryRetries))
		} else {
			plan.DeliveryRetries = types.Int64Null()
		}
		if echoResp.CreateWebSubHubNode.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateWebSubHubNode.Description)
		} else {
			plan.Description = types.StringNull()
		}
		plan.Endpoint = types.StringValue(echoResp.CreateWebSubHubNode.Endpoint)
		plan.Id = types.StringValue(echoResp.CreateWebSubHubNode.Name)
		if echoResp.CreateWebSubHubNode.InlineApiAuthenticator != nil {
			plan.InlineApiAuthenticator = types.StringValue(*echoResp.CreateWebSubHubNode.InlineApiAuthenticator)
		} else {
			plan.InlineApiAuthenticator = types.StringNull()
		}
		if echoResp.CreateWebSubHubNode.LoggingLevel != nil {
			plan.LoggingLevel = types.StringValue(string(*echoResp.CreateWebSubHubNode.LoggingLevel))
		} else {
			plan.LoggingLevel = types.StringNull()
		}
		if echoResp.CreateWebSubHubNode.ManagedApiAuthenticator != nil {
			plan.ManagedApiAuthenticator = types.StringValue(echoResp.CreateWebSubHubNode.ManagedApiAuthenticator.Name)
		} else {
			plan.ManagedApiAuthenticator = types.StringNull()
		}
		plan.MaxLeaseSeconds = types.Int64Value(int64(echoResp.CreateWebSubHubNode.MaxLeaseSeconds))
		plan.Name = types.StringValue(echoResp.CreateWebSubHubNode.Name)
		plan.ReceiveMessageType = types.StringValue(echoResp.CreateWebSubHubNode.ReceiveMessageType.Name)
		if len(echoResp.CreateWebSubHubNode.Requirements) > 0 {
			elems := []attr.Value{}
			for _, req := range echoResp.CreateWebSubHubNode.Requirements {
				elems = append(elems, types.StringValue(req))
			}
			plan.Requirements, diags = types.SetValue(types.StringType, elems)
			if diags != nil && diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
		} else {
			plan.Requirements = types.SetNull(types.StringType)
		}
		plan.SignatureAlgorithm = types.StringValue(string(echoResp.CreateWebSubHubNode.SignatureAlgorithm))
		if echoResp.CreateWebSubHubNode.SubscriptionSecurity != nil {
			plan.SubscriptionSecurity = types.StringValue(string(*echoResp.CreateWebSubHubNode.SubscriptionSecurity))
		} else {
			plan.SubscriptionSecurity = types.StringNull()
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WebSubHubNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webSubHubNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting WebSubHubNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *WebSubHubNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *WebSubHubNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_sub_hub_node"
}

func (r *WebSubHubNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		state webSubHubNodeModel
	)

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading WebSubHubNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeWebSubHubNode:
			if node.Config != nil {
				state.Config = common.ConfigValue(*node.Config)
			} else {
				state.Config = common.ConfigNull()
			}
			state.DefaultLeaseSeconds = types.Int64Value(int64(node.DefaultLeaseSeconds))
			if node.DeliveryRetries != nil {
				state.DeliveryRetries = types.Int64Value(int64(*node.DeliveryRetries))
			} else {
				state.DeliveryRetries = types.Int64Null()
			}
			if node.Description != nil {
				state.Description = types.StringValue(*node.Description)
			} else {
				state.Description = types.StringNull()
			}
			state.Endpoint = types.StringValue(node.Endpoint)
			state.Id = types.StringValue(node.Name)
			if node.InlineApiAuthenticator != nil {
				state.InlineApiAuthenticator = types.StringValue(*node.InlineApiAuthenticator)
			} else {
				state.InlineApiAuthenticator = types.StringNull()
			}
			if node.LoggingLevel != nil {
				state.LoggingLevel = types.StringValue(string(*node.LoggingLevel))
			} else {
				state.LoggingLevel = types.StringNull()
			}
			if node.ManagedApiAuthenticator != nil {
				state.ManagedApiAuthenticator = types.StringValue(node.ManagedApiAuthenticator.Name)
			} else {
				state.ManagedApiAuthenticator = types.StringNull()
			}
			state.MaxLeaseSeconds = types.Int64Value(int64(node.MaxLeaseSeconds))
			state.Name = types.StringValue(node.Name)
			state.ReceiveMessageType = types.StringValue(node.ReceiveMessageType.Name)
			if len(node.Requirements) > 0 {
				elems := []attr.Value{}
				for _, req := range node.Requirements {
					elems = append(elems, types.StringValue(req))
				}
				state.Requirements, diags = types.SetValue(types.StringType, elems)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			} else {
				state.Requirements = types.SetNull(types.StringType)
			}
			state.SignatureAlgorithm = types.StringValue(string(node.SignatureAlgorithm))
			if node.SubscriptionSecurity != nil {
				state.SubscriptionSecurity = types.StringValue(string(*node.SubscriptionSecurity))
			} else {
				state.SubscriptionSecurity = types.StringNull()
			}
		default:
			resp.Diagnostics.AddError(
				"Expected WebSubHubNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WebSubHubNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config": schema.StringAttribute{
				CustomType:          common.ConfigType{},
				MarkdownDescription: "The config, in JSON object format (i.e. - dict, map).",
				Optional:            true,
				Sensitive:           true,
			},
			"default_lease_seconds": schema.Int64Attribute{
				Computed: true,
				MarkdownDescription: "The lease duration to apply to subscription requests that do not specify hub.lease_seconds." +
					" Defaults to `864000`." +
					" Changes will only apply to new subscriptions.",
				Optional:   true,
				Validators: []validator.Int64{int64validator.AtLeast(300)},
			},
			"delivery_retries": schema.Int64Attribute{
				Computed: true,
				MarkdownDescription: "The number of times to attempt delivery to a subscription." +
					" If not provided, the subscriptions will attempt to deliver a message for 7 days." +
					" Changes will only apply to new subscriptions.",
				Optional:   true,
				Validators: []validator.Int64{int64validator.AtLeast(0)},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description.",
				Optional:            true,
			},
			"endpoint": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The WebSubHub endpoint to give to subscribers." +
					" Accepts POST calls using the WebSub protocol for subscriptions.",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"inline_api_authenticator": schema.StringAttribute{
				MarkdownDescription: "A Python code string that contains a single top-level function definition." +
					" This function must have the signature `(*, context, request, **kwargs)` and return" +
					" `None` or a tuple containing an `AuthCredentials` and `BaseUser` (or subclasses)." +
					" Mutually exclusive with `managedApiAuthenticator`.",
				Optional: true,
			},
			"logging_level": schema.StringAttribute{
				MarkdownDescription: "The logging level. One of `DEBUG`, `ERROR`, `INFO`, `WARNING`. Defaults to `INFO`.",
				Optional:            true,
				Validators:          []validator.String{common.LogLevelValidator},
			},
			"managed_api_authenticator": schema.StringAttribute{
				MarkdownDescription: "The managedApiAuthenticator. Mutually exclusive with the `inlineApiAuthenticator`.",
				Optional:            true,
			},
			"max_lease_seconds": schema.Int64Attribute{
				Computed: true,
				MarkdownDescription: "The maximum lease duration for a subscription." +
					" Defaults to `864000`." +
					" Changes will only apply to new subscriptions.",
				Optional:   true,
				Validators: []validator.Int64{int64validator.AtLeast(300)},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Node. Must be unique within the Tenant.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
				Validators:          common.FunctionNodeNameValidators,
			},
			"requirements": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format.",
				Optional:            true,
				Validators:          []validator.Set{common.RequirementsValidator},
			},
			"receive_message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Will always be 'echo.websub'",
			},
			"signature_algorithm": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The WebSub signature algorithm used by hub subscriptions when the subscription provides a secret." +
					" One of `sha1`, `sha256`, `sha384`, or `sha512`." +
					" Defaults to `sha1`." +
					" Changes will apply to all existing and new subscriptions.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(api.WebSubSignatureAlgorithmSha1),
						string(api.WebSubSignatureAlgorithmSha256),
						string(api.WebSubSignatureAlgorithmSha384),
						string(api.WebSubSignatureAlgorithmSha512),
					),
				},
			},
			"subscription_security": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The security requirements the hub is enforcing on subscription requests." +
					" One of `https`, `httpsandsecret`, or `secret`. Null indicates no enforced subscription security." +
					" Changes will only apply to new subscription requests.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(api.WebSubSubscriptionSecurityHttps),
						string(api.WebSubSubscriptionSecurityHttpsandsecret),
						string(api.WebSubSubscriptionSecuritySecret),
					),
				},
			},
		},
		MarkdownDescription: "[WebSubHubNodes](https://docs.echo.stream/docs/websub-hub) implement the W3C [WebSub](https://www.w3.org/TR/websub/) Hub feature." +
			" They accept echo.websub messages that contain content that requires publishing to subscribers.",
	}
}

func (r *WebSubHubNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webSubHubNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config                  *string
		defaultLeaseSeconds     *int
		deliveryRetries         *int
		description             *string
		diags                   diag.Diagnostics
		inlineApiAuthenticator  *string
		loggingLevel            *api.LogLevel
		managedApiAuthenticator *string
		maxLeaseSeconds         *int
		requirements            []string
		signatureAlgorithm      *api.WebSubSignatureAlgorithm
		subscriptionSecurity    *api.WebSubSubscriptionSecurity
	)
	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.DefaultLeaseSeconds.IsNull() || plan.DefaultLeaseSeconds.IsUnknown()) {
		temp := int(plan.DefaultLeaseSeconds.ValueInt64())
		defaultLeaseSeconds = &temp
	}
	if !(plan.DeliveryRetries.IsNull() || plan.DeliveryRetries.IsUnknown()) {
		temp := int(plan.DeliveryRetries.ValueInt64())
		deliveryRetries = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.InlineApiAuthenticator.IsNull() || plan.InlineApiAuthenticator.IsUnknown()) {
		temp := plan.InlineApiAuthenticator.ValueString()
		inlineApiAuthenticator = &temp
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		temp := plan.LoggingLevel.ValueString()
		loggingLevel = (*api.LogLevel)(&temp)
	}
	if !(plan.ManagedApiAuthenticator.IsNull() || plan.ManagedApiAuthenticator.IsUnknown()) {
		temp := plan.ManagedApiAuthenticator.ValueString()
		managedApiAuthenticator = &temp
	}
	if !(plan.MaxLeaseSeconds.IsNull() || plan.MaxLeaseSeconds.IsUnknown()) {
		temp := int(plan.MaxLeaseSeconds.ValueInt64())
		maxLeaseSeconds = &temp
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.SignatureAlgorithm.IsNull() || plan.SignatureAlgorithm.IsUnknown()) {
		temp := api.WebSubSignatureAlgorithm(plan.SignatureAlgorithm.ValueString())
		signatureAlgorithm = &temp
	}
	if !(plan.SubscriptionSecurity.IsNull() || plan.SubscriptionSecurity.IsUnknown()) {
		temp := api.WebSubSubscriptionSecurity(plan.SubscriptionSecurity.ValueString())
		subscriptionSecurity = &temp
	}

	if echoResp, err := api.UpdateWebSubHubNode(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		r.data.Tenant,
		config,
		defaultLeaseSeconds,
		deliveryRetries,
		description,
		inlineApiAuthenticator,
		loggingLevel,
		managedApiAuthenticator,
		maxLeaseSeconds,
		requirements,
		signatureAlgorithm,
		subscriptionSecurity,
	); err != nil {
		resp.Diagnostics.AddError("Error updating WebSubHubNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find WebSubHubNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.ValueString()))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateWebSubHubNodeGetNodeWebSubHubNode:
			if node.Update.Config != nil {
				plan.Config = common.ConfigValue(*node.Update.Config)
			} else {
				plan.Config = common.ConfigNull()
			}
			plan.DefaultLeaseSeconds = types.Int64Value(int64(node.Update.DefaultLeaseSeconds))
			if node.Update.DeliveryRetries != nil {
				plan.DeliveryRetries = types.Int64Value(int64(*node.Update.DeliveryRetries))
			} else {
				plan.DeliveryRetries = types.Int64Null()
			}
			if node.Update.Description != nil {
				plan.Description = types.StringValue(*node.Update.Description)
			} else {
				plan.Description = types.StringNull()
			}
			plan.Endpoint = types.StringValue(node.Update.Endpoint)
			plan.Id = types.StringValue(node.Update.Name)
			if node.Update.InlineApiAuthenticator != nil {
				plan.InlineApiAuthenticator = types.StringValue(*node.Update.InlineApiAuthenticator)
			} else {
				plan.InlineApiAuthenticator = types.StringNull()
			}
			if node.Update.LoggingLevel != nil {
				plan.LoggingLevel = types.StringValue(string(*node.Update.LoggingLevel))
			} else {
				plan.LoggingLevel = types.StringNull()
			}
			if node.Update.ManagedApiAuthenticator != nil {
				plan.ManagedApiAuthenticator = types.StringValue(node.Update.ManagedApiAuthenticator.Name)
			} else {
				plan.ManagedApiAuthenticator = types.StringNull()
			}
			plan.MaxLeaseSeconds = types.Int64Value(int64(node.Update.MaxLeaseSeconds))
			plan.Name = types.StringValue(node.Update.Name)
			plan.ReceiveMessageType = types.StringValue(node.Update.ReceiveMessageType.Name)
			if len(node.Update.Requirements) > 0 {
				elems := []attr.Value{}
				for _, req := range node.Update.Requirements {
					elems = append(elems, types.StringValue(req))
				}
				plan.Requirements, diags = types.SetValue(types.StringType, elems)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			} else {
				plan.Requirements = types.SetNull(types.StringType)
			}
			plan.SignatureAlgorithm = types.StringValue(string(node.Update.SignatureAlgorithm))
			if node.Update.SubscriptionSecurity != nil {
				plan.SubscriptionSecurity = types.StringValue(string(*node.Update.SubscriptionSecurity))
			} else {
				plan.SubscriptionSecurity = types.StringNull()
			}
		default:
			resp.Diagnostics.AddError(
				"Expected WebSubHubNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
