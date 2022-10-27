package message_type

import (
	"context"
	"fmt"
	"time"

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

	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		readme = &plan.Readme.Value
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		requirements = make([]string, len(plan.Requirements.Elems))
		resp.Diagnostics.Append(plan.Requirements.ElementsAs(ctx, &requirements, false)...)
		if resp.Diagnostics.HasError() {
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
		return
	}

	plan.Auditor = types.String{Value: echoResp.CreateMessageType.Auditor}
	plan.BitmapperTemplate = types.String{Value: echoResp.CreateMessageType.BitmapperTemplate}
	plan.Description = types.String{Value: echoResp.CreateMessageType.Description}
	plan.Id = types.String{Value: echoResp.CreateMessageType.Name}
	plan.InUse = types.Bool{Value: echoResp.CreateMessageType.InUse}
	plan.Name = types.String{Value: echoResp.CreateMessageType.Name}
	plan.ProcessorTemplate = types.String{Value: echoResp.CreateMessageType.ProcessorTemplate}
	if echoResp.CreateMessageType.Readme != nil {
		plan.Readme = types.String{Value: *echoResp.CreateMessageType.Readme}
	} else {
		plan.Readme = types.String{Null: true}
	}
	plan.Requirements = types.Set{ElemType: types.StringType}
	if len(echoResp.CreateMessageType.Requirements) > 0 {
		for _, req := range echoResp.CreateMessageType.Requirements {
			plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
		}
	} else {
		plan.Requirements.Null = true
	}
	plan.SampleMessage = types.String{Value: echoResp.CreateMessageType.SampleMessage}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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
		return
	}

	time.Sleep(2 * time.Second)
}

func (r *MessageTypeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: resourceMessageTypeSchema(),
		MarkdownDescription: "A specific [MessageType](https://docs.echo.stream/docs/message-types) in the Tenant. " +
			"All messages sent or received must be loosely associated (via Node and Edge typing) with a MessageType.",
	}, nil
}

func (r *MessageTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *MessageTypeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_message_type"
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
	if !prevent_destroy {
		var plan messageTypeModel

		// Read Terraform plan data into the model
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if resp.Diagnostics.HasError() {
			return
		}

		prevent_destroy = !plan.Name.Equal(state.Name)
	}

	if prevent_destroy {
		resp.Diagnostics.AddError("Cannot destroy MessageType", fmt.Sprintf("MessageType %s is in use and may not be destroyed", state.Name.Value))
	}
}

func (r *MessageTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state messageTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, system, err := readMessageType(ctx, r.data.Client, state.Name.Value, r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error reading MessageType", err.Error())
		return
	} else if system {
		resp.Diagnostics.AddError("Invalid MessageType", "Cannot import resource for system MessageType")
		return
	} else if data == nil {
		resp.State.RemoveResource(ctx)
		return
	} else {
		state = *data
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MessageTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan messageTypeModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

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
	if !(plan.Auditor.IsNull() || plan.Auditor.IsUnknown()) {
		auditor = &plan.Auditor.Value
	}
	if !(plan.BitmapperTemplate.IsNull() || plan.BitmapperTemplate.IsUnknown()) {
		bitmapperTemplate = &plan.BitmapperTemplate.Value
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		description = &plan.Description.Value
	}
	if !(plan.ProcessorTemplate.IsNull() || plan.ProcessorTemplate.IsUnknown()) {
		processorTemplate = &plan.ProcessorTemplate.Value
	}
	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		readme = &plan.Readme.Value
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		requirements = make([]string, len(plan.Requirements.Elems))
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.SampleMessage.IsNull() || plan.SampleMessage.IsUnknown()) {
		sampleMessage = &plan.SampleMessage.Value
	}

	echoResp, err := api.UpdateMessageType(
		ctx,
		r.data.Client,
		plan.Name.Value,
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
		resp.Diagnostics.AddError("Error updating Message Type", err.Error())
		return
	}

	plan.Auditor = types.String{Value: echoResp.GetMessageType.Update.Auditor}
	plan.BitmapperTemplate = types.String{Value: echoResp.GetMessageType.Update.BitmapperTemplate}
	plan.Description = types.String{Value: echoResp.GetMessageType.Update.Description}
	plan.Id = types.String{Value: echoResp.GetMessageType.Update.Name}
	plan.InUse = types.Bool{Value: echoResp.GetMessageType.Update.InUse}
	plan.Name = types.String{Value: echoResp.GetMessageType.Update.Name}
	plan.ProcessorTemplate = types.String{Value: echoResp.GetMessageType.Update.ProcessorTemplate}
	if echoResp.GetMessageType.Update.Readme != nil {
		plan.Readme = types.String{Value: *echoResp.GetMessageType.Update.Readme}
	} else {
		plan.Readme = types.String{Null: true}
	}
	plan.Requirements = types.Set{ElemType: types.StringType}
	if len(echoResp.GetMessageType.Update.Requirements) > 0 {
		for _, req := range echoResp.GetMessageType.Update.Requirements {
			plan.Requirements.Elems = append(plan.Requirements.Elems, types.String{Value: req})
		}
	} else {
		plan.Requirements.Null = true
	}
	plan.SampleMessage = types.String{Value: echoResp.GetMessageType.Update.SampleMessage}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
