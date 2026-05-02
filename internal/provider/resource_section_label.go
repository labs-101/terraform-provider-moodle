package provider

import (
	"context"
	"fmt"
	"terraform-moodle-provider/internal/moodle"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &sectionLabelResource{}
	_ resource.ResourceWithConfigure = &sectionLabelResource{}
)

func NewSectionLabelResource() resource.Resource {
	return &sectionLabelResource{}
}

type sectionLabelResource struct {
	client *moodle.MoodleClient
}

type sectionLabelResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	CourseID          types.Int64  `tfsdk:"course_id"`
	Section        types.Int64  `tfsdk:"section"`
	Intro             types.String `tfsdk:"intro"`
	Name              types.String `tfsdk:"name"`
	PreviousElementId types.Int64  `tfsdk:"previous_element_id"`
	Visible           types.Int64  `tfsdk:"visible"`
}

func (r *sectionLabelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_section_label"
}

func (r *sectionLabelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*moodle.MoodleClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider type",
			fmt.Sprintf("Expected *MoodleClient, got: %T. Please report this issue.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *sectionLabelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a Label (Textfeld) activity in a Moodle course section. Labels display HTML text directly in the course page without requiring a click-through.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The Course Module ID (cmID) of the created Label.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course to which the Label is added.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"section": schema.Int64Attribute{
				Required:    true,
				Description: "The section number (0-based) in which the Label is placed.",
			},
			"intro": schema.StringAttribute{
				Required:    true,
				Description: "HTML content of the label that is displayed directly in the course section.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Internal name of the label (not displayed to students). Defaults to an empty string.",
			},
			"previous_element_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Course Module ID of the element before which this label should be inserted. Use 0 to append at the end of the section.",
			},
			"visible": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Visibility of the label (1 = visible, 0 = hidden). Default: 1.",
			},
		},
	}
}

func (r *sectionLabelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sectionLabelResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	previousElementId := int64(0)
	if !plan.PreviousElementId.IsNull() && !plan.PreviousElementId.IsUnknown() {
		previousElementId = plan.PreviousElementId.ValueInt64()
	}

	cmID, err := r.client.CreateLabel(
		plan.CourseID.ValueInt64(),
		plan.Section.ValueInt64(),
		plan.Intro.ValueString(),
		name,
		previousElementId,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating label", err.Error())
		return
	}

	plan.ID = types.Int64Value(cmID)
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		plan.Name = types.StringValue("")
	}
	if plan.PreviousElementId.IsNull() || plan.PreviousElementId.IsUnknown() {
		plan.PreviousElementId = types.Int64Value(previousElementId)
	}
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		plan.Visible = types.Int64Value(1)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sectionLabelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sectionLabelResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	module, err := r.client.GetCourseModule(state.CourseID.ValueInt64(), state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading label", err.Error())
		return
	}

	if module == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(module.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sectionLabelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sectionLabelResourceModel
	var state sectionLabelResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	visible := int64(1)
	if !plan.Visible.IsNull() && !plan.Visible.IsUnknown() {
		visible = plan.Visible.ValueInt64()
	}

	previousElementId := int64(0)
	if !plan.PreviousElementId.IsNull() && !plan.PreviousElementId.IsUnknown() {
		previousElementId = plan.PreviousElementId.ValueInt64()
	}

	err := r.client.UpdateLabel(
		state.CourseID.ValueInt64(),
		state.ID.ValueInt64(),
		plan.Intro.ValueString(),
		plan.Name.ValueString(),
		plan.Section.ValueInt64(),
		previousElementId,
		visible,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating label", err.Error())
		return
	}

	plan.ID = state.ID
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		plan.Name = types.StringValue("")
	}
	if plan.PreviousElementId.IsNull() || plan.PreviousElementId.IsUnknown() {
		plan.PreviousElementId = types.Int64Value(previousElementId)
	}
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		plan.Visible = types.Int64Value(visible)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sectionLabelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sectionLabelResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLabel(state.CourseID.ValueInt64(), state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting label", err.Error())
	}
}
