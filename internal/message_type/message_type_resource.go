package message_type

import (
	"context"
	"fmt"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/api"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
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
	_ resource.ResourceWithConfigure   = &MessageTypeResource{}
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
		diags        diag.Diagnostics
		readme       *string
		requirements []string
	)

	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		temp := plan.Readme.ValueString()
		readme = &temp
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		resp.Diagnostics.Append(plan.Requirements.ElementsAs(ctx, &requirements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	echoResp, err := api.CreateMessageType(
		ctx,
		r.data.Client,
		plan.Auditor.ValueString(),
		plan.BitmapperTemplate.ValueString(),
		plan.Description.ValueString(),
		plan.Name.ValueString(),
		plan.ProcessorTemplate.ValueString(),
		plan.SampleMessage.ValueString(),
		r.data.Tenant,
		readme,
		requirements,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Message Type", err.Error())
		return
	}

	plan.Auditor = types.StringValue(echoResp.CreateMessageType.Auditor)
	plan.BitmapperTemplate = types.StringValue(echoResp.CreateMessageType.BitmapperTemplate)
	plan.Description = types.StringValue(echoResp.CreateMessageType.Description)
	plan.Id = types.StringValue(echoResp.CreateMessageType.Name)
	plan.InUse = types.BoolValue(echoResp.CreateMessageType.InUse)
	plan.Name = types.StringValue(echoResp.CreateMessageType.Name)
	plan.ProcessorTemplate = types.StringValue(echoResp.CreateMessageType.ProcessorTemplate)
	if echoResp.CreateMessageType.Readme != nil {
		plan.Readme = types.StringValue(*echoResp.CreateMessageType.Readme)
	} else {
		plan.Readme = types.StringNull()
	}
	if len(echoResp.CreateMessageType.Requirements) > 0 {
		elems := []attr.Value{}
		for _, req := range echoResp.CreateMessageType.Requirements {
			elems = append(elems, types.StringValue(req))
		}
		plan.Requirements, diags = types.SetValue(types.StringType, elems)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
	} else {
		plan.Requirements = types.SetNull(types.StringType)
	}
	plan.SampleMessage = types.StringValue(echoResp.CreateMessageType.SampleMessage)

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

	if _, err := api.DeleteMessageType(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); err != nil {
		resp.Diagnostics.AddError("Error deleting Message Type", err.Error())
		return
	}

	time.Sleep(2 * time.Second)
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
	if !state.InUse.ValueBool() {
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
		resp.Diagnostics.AddError("Cannot destroy MessageType", fmt.Sprintf("MessageType %s is in use and may not be destroyed", state.Name.ValueString()))
	}
}

func (r *MessageTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state messageTypeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data, system, diags := readMessageType(ctx, r.data.Client, state.Name.ValueString(), r.data.Tenant); diags.HasError() {
		resp.Diagnostics.Append(diags...)
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

func (r *MessageTypeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"auditor": schema.StringAttribute{
				MarkdownDescription: "A Python code string that contains a single top-level function definition." +
					" This function must have the signature `(*, message, **kwargs)` where" +
					" message is a string and must return a flat dictionary.",
				Required: true,
			},
			"bitmapper_template": schema.StringAttribute{
				MarkdownDescription: " A Python code string that contains a single top-level function definition." +
					" This function is used as a template when creating custom routing rules in" +
					" RouterNodes that use this MessageType. This function must have the signature" +
					" `(*, context, message, source, **kwargs)` and return an integer.",
				Required: true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"in_use": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "True if this is used by other resources.",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the MessageType.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:          append(messageTypeNameValidators, common.NotSystemNameValidator),
			},
			"processor_template": schema.StringAttribute{
				MarkdownDescription: " A Python code string that contains a single top-leve function definition." +
					" This function is used as a template when creating custom processing in" +
					" ProcessorNodes that use this MessageType. This function must have the signature" +
					" `(*, context, message, source, **kwargs)` and return `None`, a string or a list of strings.",
				Required: true,
			},
			"readme": schema.StringAttribute{
				MarkdownDescription: "README in MarkDown format.",
				Optional:            true,
			},
			"requirements": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format.",
				Optional:            true,
				Validators:          []validator.Set{common.RequirementsValidator},
			},
			"sample_message": schema.StringAttribute{
				MarkdownDescription: "A sample message.",
				Required:            true,
			},
		},
		MarkdownDescription: "A specific [MessageType](https://docs.echo.stream/docs/message-types) in the Tenant. " +
			"All messages sent or received must be loosely associated (via Node and Edge typing) with a MessageType.",
	}
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
		diags             diag.Diagnostics
		processorTemplate *string
		readme            *string
		requirements      []string
		sampleMessage     *string
	)
	if !(plan.Auditor.IsNull() || plan.Auditor.IsUnknown()) {
		temp := plan.Auditor.ValueString()
		auditor = &temp
	}
	if !(plan.BitmapperTemplate.IsNull() || plan.BitmapperTemplate.IsUnknown()) {
		temp := plan.BitmapperTemplate.ValueString()
		bitmapperTemplate = &temp
	}
	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		temp := plan.Description.ValueString()
		description = &temp
	}
	if !(plan.ProcessorTemplate.IsNull() || plan.ProcessorTemplate.IsUnknown()) {
		temp := plan.ProcessorTemplate.ValueString()
		processorTemplate = &temp
	}
	if !(plan.Readme.IsNull() || plan.Readme.IsUnknown()) {
		temp := plan.Readme.ValueString()
		readme = &temp
	}
	if !(plan.Requirements.IsNull() || plan.Requirements.IsUnknown()) {
		diags := plan.Requirements.ElementsAs(ctx, &requirements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if !(plan.SampleMessage.IsNull() || plan.SampleMessage.IsUnknown()) {
		temp := plan.SampleMessage.ValueString()
		sampleMessage = &temp
	}

	echoResp, err := api.UpdateMessageType(
		ctx,
		r.data.Client,
		plan.Name.ValueString(),
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

	plan.Auditor = types.StringValue(echoResp.GetMessageType.Update.Auditor)
	plan.BitmapperTemplate = types.StringValue(echoResp.GetMessageType.Update.BitmapperTemplate)
	plan.Description = types.StringValue(echoResp.GetMessageType.Update.Description)
	plan.Id = types.StringValue(echoResp.GetMessageType.Update.Name)
	plan.InUse = types.BoolValue(echoResp.GetMessageType.Update.InUse)
	plan.Name = types.StringValue(echoResp.GetMessageType.Update.Name)
	plan.ProcessorTemplate = types.StringValue(echoResp.GetMessageType.Update.ProcessorTemplate)
	if echoResp.GetMessageType.Update.Readme != nil {
		plan.Readme = types.StringValue(*echoResp.GetMessageType.Update.Readme)
	} else {
		plan.Readme = types.StringNull()
	}
	if len(echoResp.GetMessageType.Update.Requirements) > 0 {
		elems := []attr.Value{}
		for _, req := range echoResp.GetMessageType.Update.Requirements {
			elems = append(elems, types.StringValue(req))
		}
		plan.Requirements, diags = types.SetValue(types.StringType, elems)
		if diags != nil && diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
	} else {
		plan.Requirements = types.SetNull(types.StringType)
	}
	plan.SampleMessage = types.StringValue(echoResp.GetMessageType.Update.SampleMessage)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
