package message_type

import (
	"context"
	"fmt"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.ResourceWithImportState = &MessageTypeResource{}
	_ resource.ResourceWithModifyPlan  = &MessageTypeResource{}
)

// MessageTypeResource defines the resource implementation.
type MessageTypeResource struct {
	data *common.ProviderData
}

func (r *MessageTypeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_message_type"
}

func (r *MessageTypeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes:          resourceMessageTypeSchema(),
		Description:         "Message Types provide a loose typing system in EchoStream",
		MarkdownDescription: "Message Types provide a loose typing system in EchoStream",
	}, nil
}

func (r *MessageTypeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MessageTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan messageTypeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		readme       *string
		requirements []string
	)
	if !plan.Readme.IsNull() {
		readme = &plan.Readme.Value
	}
	if !plan.Requirements.IsNull() {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	echoResp, err := api.CreateMessageType(
		ctx,
		r.data.Client,
		plan.Auditor.Value,
		plan.BitmapperTemplate.Value,
		plan.Description.Value,
		plan.Name.Value,
		plan.ProcessorTemplate.Value,
		plan.SampleMessage.Value,
		r.data.Tenant,
		readme,
		requirements,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Message Type", err.Error())
	}

	plan.Auditor = types.String{Value: echoResp.CreateMessageType.Auditor}
	plan.BitmapperTemplate = types.String{Value: echoResp.CreateMessageType.BitmapperTemplate}
	plan.Description = types.String{Value: echoResp.CreateMessageType.Description}
	plan.InUse = types.Bool{Value: echoResp.CreateMessageType.InUse}
	plan.Name = types.String{Value: echoResp.CreateMessageType.Name}
	plan.ProcessorTemplate = types.String{Value: echoResp.CreateMessageType.ProcessorTemplate}
	if echoResp.CreateMessageType.Readme != nil {
		plan.Readme = types.String{Value: *echoResp.CreateMessageType.Readme}
	} else {
		plan.Readme = types.String{Null: true}
	}
	if len(echoResp.CreateMessageType.Requirements) > 0 {
		plan.Requirements = types.Set{ElemType: types.StringType}
		for _, req := range echoResp.CreateMessageType.Requirements {
			plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
		}
	} else {
		plan.Requirements = types.Set{ElemType: types.StringType, Null: true}
	}
	plan.SampleMessage = types.String{Value: echoResp.CreateMessageType.SampleMessage}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MessageTypeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var state messageTypeModel

	// If the entire state is null, resource is being created.
	if req.State.Raw.IsNull() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If the MessageType is not in use it can be destroyed at will.
	if !state.InUse.Value {
		return
	}

	// If the entire plan is null, the resource is planned for destruction.
	prevent_destroy := req.Plan.Raw.IsNull()
	detail := fmt.Sprintf("MessageType %s is in use and may not be destroyed", state.Name.Value)
	if !prevent_destroy {
		var plan messageTypeModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		prevent_destroy = !plan.Name.Equal(state.Name)
		detail = fmt.Sprintf("MessageType %s is in use and may not be replaced", state.Name.Value)
	}

	if prevent_destroy {
		resp.Diagnostics.AddError("Cannot destroy Message Type", detail)
	}
}

func (r *MessageTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state messageTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if system, err := readMessageType(ctx, r.data.Client, state.Name.Value, r.data.Tenant, &state); err != nil {
		resp.Diagnostics.AddError("Error reading MessageType", err.Error())
		return
	} else if system {
		resp.Diagnostics.AddError("Invalid MessageType", "Cannot create resource for system MessageType")
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MessageTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state *messageTypeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		auditor           *string
		bitmapperTemplate *string
		description       *string
		processorTemplate *string
		readme            *string
		requirements      []string
		sampleMessage     *string
	)
	if !state.Auditor.IsNull() {
		auditor = &state.Auditor.Value
	}
	if !state.BitmapperTemplate.IsNull() {
		bitmapperTemplate = &state.BitmapperTemplate.Value
	}
	if !state.Description.IsNull() {
		description = &state.Description.Value
	}
	if !state.ProcessorTemplate.IsNull() {
		processorTemplate = &state.ProcessorTemplate.Value
	}
	if !state.Readme.IsNull() {
		readme = &state.Readme.Value
	}
	if !state.Requirements.IsNull() {
		requirements = make([]string, len(state.Requirements.Elems))
		diags := state.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !state.SampleMessage.IsNull() {
		sampleMessage = &state.SampleMessage.Value
	}

	echoResp, err := api.UpdateMessageType(
		ctx,
		r.data.Client,
		state.Name.Value,
		r.data.Tenant,
		auditor,
		bitmapperTemplate,
		description,
		processorTemplate,
		readme,
		requirements,
		sampleMessage,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Message Type", err.Error())
	}

	state.Auditor = types.String{Value: echoResp.GetMessageType.Update.Auditor}
	state.BitmapperTemplate = types.String{Value: echoResp.GetMessageType.Update.BitmapperTemplate}
	state.Description = types.String{Value: echoResp.GetMessageType.Update.Description}
	state.InUse = types.Bool{Value: echoResp.GetMessageType.Update.InUse}
	state.Name = types.String{Value: echoResp.GetMessageType.Update.Name}
	state.ProcessorTemplate = types.String{Value: echoResp.GetMessageType.Update.ProcessorTemplate}
	if echoResp.GetMessageType.Update.Readme != nil {
		state.Readme = types.String{Value: *echoResp.GetMessageType.Update.Readme}
	} else {
		state.Readme = types.String{Null: true}
	}
	if len(echoResp.GetMessageType.Update.Requirements) > 0 {
		state.Requirements = types.Set{ElemType: types.StringType}
		for _, req := range echoResp.GetMessageType.Update.Requirements {
			state.Requirements.Elems = append(state.Requirements.Elems, types.String{Value: req})
		}
	} else {
		state.Requirements = types.Set{ElemType: types.StringType, Null: true}
	}
	state.SampleMessage = types.String{Value: echoResp.GetMessageType.Update.SampleMessage}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MessageTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state messageTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := api.DeleteMessageType(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting Message Type", err.Error())
	}
}

func (r *MessageTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
