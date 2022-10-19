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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithImportState      = &BitmapRouterNodeResource{}
	_ resource.ResourceWithConfigValidators = &BitmapRouterNodeResource{}
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
		resourcevalidator.ExactlyOneOf(
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
		inlineBitmapper  *string
		loggingLevel     *api.LogLevel
		managedBitmapper *string
		requirements     []string
		routeTable       *string
	)

	if !(plan.Config.IsNull() || plan.Config.IsUnknown()) {
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.InlineBitmapper.IsNull() || plan.InlineBitmapper.IsUnknown()) {
		inlineBitmapper = &plan.InlineBitmapper.Value
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		loggingLevel = (*api.LogLevel)(&plan.LoggingLevel.Value)
	}
	if !(plan.ManagedBitmapper.IsNull() || plan.ManagedBitmapper.IsUnknown()) {
		managedBitmapper = &plan.ManagedBitmapper.Value
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.RouteTable.IsNull() || plan.RouteTable.IsUnknown()) {
		rt := make(map[string][]string, len(plan.RouteTable.Elems))
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
		plan.Name.Value,
		plan.ReceiveMessageType.Value,
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
			plan.Config = common.Config{Value: *echoResp.CreateBitmapRouterNode.Config}
		} else {
			plan.Config = common.Config{Null: true}
		}
		if echoResp.CreateBitmapRouterNode.Description != nil {
			plan.Description = types.String{Value: *echoResp.CreateBitmapRouterNode.Description}
		} else {
			plan.Description = types.String{Null: true}
		}
		if echoResp.CreateBitmapRouterNode.InlineBitmapper != nil {
			plan.InlineBitmapper = types.String{Value: *echoResp.CreateBitmapRouterNode.InlineBitmapper}
		} else {
			plan.InlineBitmapper = types.String{Null: true}
		}
		if echoResp.CreateBitmapRouterNode.LoggingLevel != nil {
			plan.LoggingLevel = types.String{Value: string(*echoResp.CreateBitmapRouterNode.LoggingLevel)}
		} else {
			plan.LoggingLevel = types.String{Null: true}
		}
		if echoResp.CreateBitmapRouterNode.ManagedBitmapper != nil {
			plan.ManagedBitmapper = types.String{Value: echoResp.CreateBitmapRouterNode.ManagedBitmapper.Name}
		} else {
			plan.ManagedBitmapper = types.String{Null: true}
		}
		plan.Name = types.String{Value: echoResp.CreateBitmapRouterNode.Name}
		plan.ReceiveMessageType = types.String{Value: echoResp.CreateBitmapRouterNode.ReceiveMessageType.Name}
		if len(echoResp.CreateBitmapRouterNode.Requirements) > 0 {
			plan.Requirements = types.Set{ElemType: types.StringType}
			for _, req := range echoResp.CreateBitmapRouterNode.Requirements {
				plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
			}
		} else {
			plan.Requirements = types.Set{ElemType: types.StringType, Null: true}
		}
		rt := map[string][]string{}
		if err := json.Unmarshal([]byte(echoResp.CreateBitmapRouterNode.RouteTable), &rt); err != nil {
			resp.Diagnostics.AddError("Error unmashalling route_table", err.Error())
			return
		} else if len(rt) > 0 {
			plan.RouteTable = types.Map{Elems: map[string]attr.Value{}, ElemType: types.SetType{ElemType: types.StringType}}
			for route_bitmap, t := range rt {
				targets := types.Set{ElemType: types.StringType}
				for _, target := range t {
					targets.Elems = append(targets.Elems, types.String{Value: target})
				}
				plan.RouteTable.Elems[route_bitmap] = targets
			}
		} else {
			plan.RouteTable = types.Map{ElemType: types.SetType{ElemType: types.StringType}, Null: true}
		}
		plan.SendMessageType = types.String{Value: echoResp.CreateBitmapRouterNode.SendMessageType.Name}
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

	if _, err := api.DeleteNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting BitmapRouterNode", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *BitmapRouterNodeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := dataSendReceiveNodeSchema()
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
		case "receive_message_type":
			attribute.Computed = false
			attribute.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
			attribute.Required = true
		}
		schema[key] = attribute
	}
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"config": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Sensitive:           true,
				Type:                common.ConfigType{},
			},
			"inline_bitmapper": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
			},
			"logging_level": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
				Validators:          []tfsdk.AttributeValidator{common.LogLevelValidator},
			},
			"managed_bitmapper": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
			},
			"requirements": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.SetType{ElemType: types.StringType},
				Validators:          []tfsdk.AttributeValidator{common.RequirementsValidator},
			},
			"route_table": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.MapType{ElemType: types.SetType{ElemType: types.StringType}},
				Validators: []tfsdk.AttributeValidator{
					mapvalidator.KeysAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^0[x|X][a-fA-F0-9]+$`),
							"Must begin with '0x' or '0X' and contain 'a-f', 'A-F' or '0-9'",
						),
					),
					mapvalidator.ValuesAre(
						setvalidator.ValuesAre(
							stringvalidator.LengthAtLeast(1),
						),
					),
				},
			},
		},
	)
	return tfsdk.Schema{
		Attributes:          schema,
		Description:         "BitmapRouterNodes allow for the routing of messages based upon message content",
		MarkdownDescription: "BitmapRouterNodes allow for the routing of messages based upon message content",
	}, nil
}

func (r *BitmapRouterNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *BitmapRouterNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bitmap_router_node"
}

func (r *BitmapRouterNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bitmapRouterNodeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if echoResp, err := api.ReadNode(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading BitmapRouterNode", err.Error())
		return
	} else if echoResp.GetNode == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.ReadNodeGetNodeBitmapRouterNode:
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
			if node.InlineBitmapper != nil {
				state.InlineBitmapper = types.String{Value: *node.InlineBitmapper}
			} else {
				state.InlineBitmapper = types.String{Null: true}
			}
			if node.LoggingLevel != nil {
				state.LoggingLevel = types.String{Value: string(*node.LoggingLevel)}
			} else {
				state.LoggingLevel = types.String{Null: true}
			}
			if node.ManagedBitmapper != nil {
				state.ManagedBitmapper = types.String{Value: node.ManagedBitmapper.Name}
			} else {
				state.ManagedBitmapper = types.String{Null: true}
			}
			state.Name = types.String{Value: node.Name}
			state.ReceiveMessageType = types.String{Value: node.ReceiveMessageType.Name}
			if len(node.Requirements) > 0 {
				state.Requirements = types.Set{ElemType: types.StringType}
				for _, req := range node.Requirements {
					state.Requirements.Elems = append(state.Requirements.Elems, types.String{Value: req})
				}
			} else {
				state.Requirements = types.Set{ElemType: types.StringType, Null: true}
			}
			rt := map[string][]string{}
			if err := json.Unmarshal([]byte(node.RouteTable), &rt); err != nil {
				resp.Diagnostics.AddError("Error unmashalling route_table", err.Error())
				return
			} else if len(rt) > 0 {
				state.RouteTable = types.Map{Elems: map[string]attr.Value{}, ElemType: types.SetType{ElemType: types.StringType}}
				for route_bitmap, t := range rt {
					targets := types.Set{ElemType: types.StringType}
					for _, target := range t {
						targets.Elems = append(targets.Elems, types.String{Value: target})
					}
					state.RouteTable.Elems[route_bitmap] = targets
				}
			} else {
				state.RouteTable = types.Map{ElemType: types.SetType{ElemType: types.StringType}, Null: true}
			}
			state.SendMessageType = types.String{Value: node.SendMessageType.Name}
		default:
			resp.Diagnostics.AddError(
				"Expected BitmapRouterNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), state.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BitmapRouterNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bitmapRouterNodeModel

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
		config = &plan.Config.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.InlineBitmapper.IsNull() || plan.InlineBitmapper.IsUnknown()) {
		inlineBitmapper = &plan.InlineBitmapper.Value
	}
	if !(plan.LoggingLevel.IsNull() || plan.LoggingLevel.IsUnknown()) {
		loggingLevel = (*api.LogLevel)(&plan.LoggingLevel.Value)
	}
	if !(plan.ManagedBitmapper.IsNull() || plan.ManagedBitmapper.IsUnknown()) {
		managedBitmapper = &plan.ManagedBitmapper.Value
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.RouteTable.IsNull() || plan.RouteTable.IsUnknown()) {
		rt := make(map[string][]string, len(plan.RouteTable.Elems))
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
		plan.Name.Value,
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
		resp.Diagnostics.AddError("Cannot find BitmapRouterNode", fmt.Sprintf("'%s' Node does not exist", plan.Name.Value))
		return
	} else {
		switch node := (*echoResp.GetNode).(type) {
		case *api.UpdateBitmapRouterNodeGetNodeBitmapRouterNode:
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
			if node.Update.InlineBitmapper != nil {
				plan.InlineBitmapper = types.String{Value: *node.Update.InlineBitmapper}
			} else {
				plan.InlineBitmapper = types.String{Null: true}
			}
			if node.Update.LoggingLevel != nil {
				plan.LoggingLevel = types.String{Value: string(*node.Update.LoggingLevel)}
			} else {
				plan.LoggingLevel = types.String{Null: true}
			}
			if node.Update.ManagedBitmapper != nil {
				plan.ManagedBitmapper = types.String{Value: node.Update.ManagedBitmapper.Name}
			} else {
				plan.ManagedBitmapper = types.String{Null: true}
			}
			plan.Name = types.String{Value: node.Update.Name}
			plan.ReceiveMessageType = types.String{Value: node.Update.ReceiveMessageType.Name}
			if len(node.Update.Requirements) > 0 {
				plan.Requirements = types.Set{ElemType: types.StringType}
				for _, req := range node.Update.Requirements {
					plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
				}
			} else {
				plan.Requirements = types.Set{ElemType: types.StringType, Null: true}
			}
			rt := map[string][]string{}
			if err := json.Unmarshal([]byte(node.Update.RouteTable), &rt); err != nil {
				resp.Diagnostics.AddError("Error unmashalling route_table", err.Error())
				return
			} else if len(rt) > 0 {
				plan.RouteTable = types.Map{Elems: map[string]attr.Value{}, ElemType: types.SetType{ElemType: types.StringType}}
				for route_bitmap, t := range rt {
					targets := types.Set{ElemType: types.StringType}
					for _, target := range t {
						targets.Elems = append(targets.Elems, types.String{Value: target})
					}
					plan.RouteTable.Elems[route_bitmap] = targets
				}
			} else {
				plan.RouteTable = types.Map{ElemType: types.SetType{ElemType: types.StringType}, Null: true}
			}
			plan.SendMessageType = types.String{Value: node.Update.SendMessageType.Name}
		default:
			resp.Diagnostics.AddError(
				"Expected BitmapRouterNode",
				fmt.Sprintf("Received '%s' for '%s'", *(*echoResp.GetNode).GetTypename(), plan.Name.Value),
			)
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
