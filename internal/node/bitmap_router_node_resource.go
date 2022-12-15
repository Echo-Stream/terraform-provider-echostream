package node

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
	_ resource.ResourceWithConfigure        = &BitmapRouterNodeResource{}
	_ resource.ResourceWithConfigValidators = &BitmapRouterNodeResource{}
	_ resource.ResourceWithImportState      = &BitmapRouterNodeResource{}
)

// BitmapRouterNodeResource defines the resource implementation.
type BitmapRouterNodeResource struct {
	data *common.ProviderData
}

type bitmapRouterNodeModel struct {
	Config             common.Config `tfsdk:"config"`
	Description        types.String  `tfsdk:"description"`
	InlineBitmapper    types.String  `tfsdk:"inline_bitmapper"`
	LoggingLevel       types.String  `tfsdk:"logging_level"`
	ManagedBitmapper   types.String  `tfsdk:"managed_bitmapper"`
	Name               types.String  `tfsdk:"name"`
	ReceiveMessageType types.String  `tfsdk:"receive_message_type"`
	Requirements       types.Set     `tfsdk:"requirements"`
	RouteTable         types.Map     `tfsdk:"route_table"`
	SendMessageType    types.String  `tfsdk:"send_message_type"`
}

func (r *BitmapRouterNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BitmapRouterNodeResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("inline_bitmapper"),
			path.MatchRoot("managed_bitmapper"),
		),
	}
}

func (r *BitmapRouterNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bitmapRouterNodeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config           *string
		description      *string
		diags            diag.Diagnostics
		inlineBitmapper  *string
		loggingLevel     *api.LogLevel
		managedBitmapper *string
		requirements     []string
		routeTable       *string
	)

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.InlineBitmapper.IsNull() || plan.InlineBitmapper.IsUnknown()) {
		temp := plan.InlineBitmapper.ValueString()
		inlineBitmapper = &temp
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		temp := plan.LoggingLevel.ValueString()
		loggingLevel = (*api.LogLevel)(&temp)
	}
	if !(plan.ManagedBitmapper.IsNull() || plan.ManagedBitmapper.IsUnknown()) {
		temp := plan.ManagedBitmapper.ValueString()
		managedBitmapper = &temp
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.RouteTable.IsNull() || plan.RouteTable.IsUnknown()) {
		rt := map[string][]string{}
		resp.Diagnostics.Append(plan.RouteTable.ElementsAs(ctx, &rt, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if b, err := json.Marshal(rt); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("route_table"), "Error marshalling to JSON", err.Error())
		} else {
			s := string(b)
			routeTable = &s
		}
	}

	if echoResp, err := api.CreateBitmapRouterNode(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		plan.ReceiveMessageType.ValueString(),
		r.data.Tenant,
		config,
		description,
		inlineBitmapper,
		loggingLevel,
		managedBitmapper,
		requirements,
		routeTable,
	); err != nil {
		resp.Diagnostics.AddError("Error creating BitmapRouterNode", err.Error())
		return
	} else {
		if echoResp.CreateBitmapRouterNode.Config != nil {
			plan.Config = common.ConfigValue(*echoResp.CreateBitmapRouterNode.Config)
		} else {
			plan.Config = common.ConfigNull()
		}
		if echoResp.CreateBitmapRouterNode.Description != nil {
			plan.Description = types.StringValue(*echoResp.CreateBitmapRouterNode.Description)
		} else {
			plan.Description = types.StringNull()
		}
		if echoResp.CreateBitmapRouterNode.InlineBitmapper != nil {
			plan.InlineBitmapper = types.StringValue(*echoResp.CreateBitmapRouterNode.InlineBitmapper)
		} else {
			plan.InlineBitmapper = types.StringNull()
		}
		if echoResp.CreateBitmapRouterNode.LoggingLevel != nil {
			plan.LoggingLevel = types.StringValue(string(*echoResp.CreateBitmapRouterNode.LoggingLevel))
		} else {
			plan.LoggingLevel = types.StringNull()
		}
		if echoResp.CreateBitmapRouterNode.ManagedBitmapper != nil {
			plan.ManagedBitmapper = types.StringValue(echoResp.CreateBitmapRouterNode.ManagedBitmapper.Name)
		} else {
			plan.ManagedBitmapper = types.StringNull()
		}
		plan.Name = types.StringValue(echoResp.CreateBitmapRouterNode.Name)
		plan.ReceiveMessageType = types.StringValue(echoResp.CreateBitmapRouterNode.ReceiveMessageType.Name)
		if len(echoResp.CreateBitmapRouterNode.Requirements) > 0 {
			elems := []attr.Value{}
			for _, req := range echoResp.CreateBitmapRouterNode.Requirements {
				elems = append(elems, types.StringValue(req))
			}
			plan.Requirements, diags = types.SetValue(types.StringType, elems)
			if diags != nil && diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
		} else {
			plan.Requirements = types.SetNull(types.StringType)
		}
		rt := map[string][]string{}
		if err := json.Unmarshal([]byte(echoResp.CreateBitmapRouterNode.RouteTable), &rt); err != nil {
			resp.Diagnostics.AddError("Error unmashalling route_table", err.Error())
			return
		} else if len(rt) > 0 {
			elems := map[string]attr.Value{}
			for route_bitmap, t := range rt {
				targets := []attr.Value{}
				for _, target := range t {
					targets = append(targets, types.StringValue(target))
				}
				elems[route_bitmap], diags = types.SetValue(types.StringType, targets)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			}
			plan.RouteTable, diags = types.MapValue(types.SetType{ElemType: types.StringType}, elems)
			if diags != nil && diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
		} else {
			plan.RouteTable = types.MapNull(types.SetType{ElemType: types.StringType})
		}
		plan.SendMessageType = types.StringValue(echoResp.CreateBitmapRouterNode.SendMessageType.Name)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BitmapRouterNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bitmapRouterNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting BitmapRouterNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *BitmapRouterNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *BitmapRouterNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bitmap_router_node"
}

func (r *BitmapRouterNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		state bitmapRouterNodeModel
	)

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading BitmapRouterNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeBitmapRouterNode:
			if node.Config != nil {
				state.Config = common.ConfigValue(*node.Config)
			} else {
				state.Config = common.ConfigNull()
			}
			if node.Description != nil {
				state.Description = types.StringValue(*node.Description)
			} else {
				state.Description = types.StringNull()
			}
			if node.InlineBitmapper != nil {
				state.InlineBitmapper = types.StringValue(*node.InlineBitmapper)
			} else {
				state.InlineBitmapper = types.StringNull()
			}
			if node.LoggingLevel != nil {
				state.LoggingLevel = types.StringValue(string(*node.LoggingLevel))
			} else {
				state.LoggingLevel = types.StringNull()
			}
			if node.ManagedBitmapper != nil {
				state.ManagedBitmapper = types.StringValue(node.ManagedBitmapper.Name)
			} else {
				state.ManagedBitmapper = types.StringNull()
			}
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
			rt := map[string][]string{}
			if err := json.Unmarshal([]byte(node.RouteTable), &rt); err != nil {
				resp.Diagnostics.AddError("Error unmashalling route_table", err.Error())
				return
			} else if len(rt) > 0 {
				elems := map[string]attr.Value{}
				for route_bitmap, t := range rt {
					targets := []attr.Value{}
					for _, target := range t {
						targets = append(targets, types.StringValue(target))
					}
					elems[route_bitmap], diags = types.SetValue(types.StringType, targets)
					if diags != nil && diags.HasError() {
						resp.Diagnostics.Append(diags...)
					}
				}
				state.RouteTable, diags = types.MapValue(types.SetType{ElemType: types.StringType}, elems)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			} else {
				state.RouteTable = types.MapNull(types.SetType{ElemType: types.StringType})
			}
			state.SendMessageType = types.StringValue(node.SendMessageType.Name)
		default:
			resp.Diagnostics.AddError(
				"Expected BitmapRouterNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BitmapRouterNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config": schema.StringAttribute{
				CustomType:          common.ConfigType{},
				MarkdownDescription: "The config, in JSON object format (i.e. - dict, map).",
				Optional:            true,
				Sensitive:           true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description.",
				Optional:            true,
			},
			"inline_bitmapper": schema.StringAttribute{
				MarkdownDescription: "A Python code string that contains a single top-level function definition." +
					"This function must have the signature `(*, context, message, source, **kwargs)`" +
					"and return an integer. Mutually exclusive with `managedBitmapper`.",
				Optional: true,
			},
			"logging_level": schema.StringAttribute{
				MarkdownDescription: "The logging level. One of `DEBUG`, `ERROR`, `INFO`, `WARNING`. Defaults to `INFO`.",
				Optional:            true,
				Validators:          []validator.String{common.LogLevelValidator},
			},
			"managed_bitmapper": schema.StringAttribute{
				MarkdownDescription: "A managed BitmapperFunction. Mutually exclusive with `inlineBitmapper`.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Node. Must be unique within the Tenant.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
				Validators:          common.FunctionNodeNameValidators,
			},
			"receive_message_type": schema.StringAttribute{
				MarkdownDescription: "The MessageType that this Node is capable of receiving.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Required:            true,
			},
			"requirements": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format.",
				Optional:            true,
				Validators:          []validator.Set{common.RequirementsValidator},
			},
			"route_table": schema.MapAttribute{
				ElementType: types.SetType{ElemType: types.StringType},
				MarkdownDescription: "The route table. A route table is a JSON object with hexidecimal (base-16) keys " +
					"(the route bitmaps - e.g. 0xF1) and a list of target Node names as the values.",
				Optional: true,
				Validators: []validator.Map{
					mapvalidator.KeysAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^0[x|X][a-fA-F0-9]+$`),
							"Must begin with '0x' or '0X' and contain 'a-f', 'A-F' or '0-9'",
						),
					),
					mapvalidator.ValueSetsAre(
						setvalidator.ValueStringsAre(
							stringvalidator.LengthAtLeast(1),
						),
					),
				},
			},
			"send_message_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The MessageType that this Node is capable of sending.",
			},
		},
		MarkdownDescription: "[BitmapRouterNodes](https://docs.echo.stream/docs/bitmap-router-node) use a bitmapper function (either " +
			"inline or managed) to construct a bitmap of truthy values for each message processed. The message bitmap is then _and_'ed with " +
			"route bitmaps. If the result of the _and_ is equal to the route bitmap then the message is sent along that route.",
	}
}

func (r *BitmapRouterNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics
		plan  bitmapRouterNodeModel
	)

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		config           *string
		description      *string
		inlineBitmapper  *string
		loggingLevel     *api.LogLevel
		managedBitmapper *string
		requirements     []string
		routeTable       *string
	)
	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		temp := plan.Config.ValueConfig()
		config = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.InlineBitmapper.IsNull() || plan.InlineBitmapper.IsUnknown()) {
		temp := plan.InlineBitmapper.ValueString()
		inlineBitmapper = &temp
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		temp := plan.LoggingLevel.ValueString()
		loggingLevel = (*api.LogLevel)(&temp)
	}
	if !(plan.ManagedBitmapper.IsNull() || plan.ManagedBitmapper.IsUnknown()) {
		temp := plan.ManagedBitmapper.ValueString()
		managedBitmapper = &temp
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.RouteTable.IsNull() || plan.RouteTable.IsUnknown()) {
		rt := map[string][]string{}
		resp.Diagnostics.Append(plan.RouteTable.ElementsAs(ctx, &rt, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if b, err := json.Marshal(rt); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("route_table"), "Error marshalling to JSON", err.Error())
		} else {
			s := string(b)
			routeTable = &s
		}
	}

	if echoResp, err := api.UpdateBitmapRouterNode(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
		r.data.Tenant,
		config,
		description,
		inlineBitmapper,
		loggingLevel,
		managedBitmapper,
		requirements,
		routeTable,
	); err != nil {
		resp.Diagnostics.AddError("Error updating BitmapRouterNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.Diagnostics.AddError("Cannot find BitmapRouterNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.ValueString()))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateBitmapRouterNodeGetNodeBitmapRouterNode:
			if node.Update.Config != nil {
				plan.Config = common.ConfigValue(*node.Update.Config)
			} else {
				plan.Config = common.ConfigNull()
			}
			if node.Update.Description != nil {
				plan.Description = types.StringValue(*node.Update.Description)
			} else {
				plan.Description = types.StringNull()
			}
			if node.Update.InlineBitmapper != nil {
				plan.InlineBitmapper = types.StringValue(*node.Update.InlineBitmapper)
			} else {
				plan.InlineBitmapper = types.StringNull()
			}
			if node.Update.LoggingLevel != nil {
				plan.LoggingLevel = types.StringValue(string(*node.Update.LoggingLevel))
			} else {
				plan.LoggingLevel = types.StringNull()
			}
			if node.Update.ManagedBitmapper != nil {
				plan.ManagedBitmapper = types.StringValue(node.Update.ManagedBitmapper.Name)
			} else {
				plan.ManagedBitmapper = types.StringNull()
			}
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
			rt := map[string][]string{}
			if err := json.Unmarshal([]byte(node.Update.RouteTable), &rt); err != nil {
				resp.Diagnostics.AddError("Error unmashalling route_table", err.Error())
				return
			} else if len(rt) > 0 {
				elems := map[string]attr.Value{}
				for route_bitmap, t := range rt {
					targets := []attr.Value{}
					for _, target := range t {
						targets = append(targets, types.StringValue(target))
					}
					elems[route_bitmap], diags = types.SetValue(types.StringType, targets)
					if diags != nil && diags.HasError() {
						resp.Diagnostics.Append(diags...)
					}
				}
				plan.RouteTable, diags = types.MapValue(types.SetType{ElemType: types.StringType}, elems)
				if diags != nil && diags.HasError() {
					resp.Diagnostics.Append(diags...)
				}
			} else {
				plan.RouteTable = types.MapNull(types.SetType{ElemType: types.StringType})
			}
			plan.SendMessageType = types.StringValue(node.Update.SendMessageType.Name)
		default:
			resp.Diagnostics.AddError(
				"Expected BitmapRouterNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.ValueString()),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
