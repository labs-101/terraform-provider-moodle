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
	_ resource.Resource              = &courseSectionResource{}
	_ resource.ResourceWithConfigure = &courseSectionResource{}
)

func NewCourseSectionResource() resource.Resource {
	return &courseSectionResource{}
}

type courseSectionResource struct {
	client *moodle.MoodleClient
}

type courseSectionResourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	CourseID types.Int64  `tfsdk:"course_id"`
	Name     types.String `tfsdk:"name"`
	Summary  types.String `tfsdk:"summary"`
	Section  types.Int64  `tfsdk:"section"`
	Visible  types.Int64  `tfsdk:"visible"`
}

func (r *courseSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_course_section"
}

func (r *courseSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*moodle.MoodleClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider type",
			fmt.Sprintf("Expected *MoodleClient, got: %T. Please report this error.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *courseSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a section in a Moodle course.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The internal database ID of the Moodle section.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course to which this section belongs.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the section.",
			},
			"summary": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The summary/description of the section (HTML is supported).",
			},
			"section": schema.Int64Attribute{
				Computed:    true,
				Description: "The section number (position) within the course, assigned by Moodle.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"visible": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Visibility of the section (1 = visible, 0 = hidden). Default: 1.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *courseSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan courseSectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 1. Add new section to course (Moodle appends it to the end)
	section, err := r.client.CreateSection(plan.CourseID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error creating section", err.Error())
		return
	}

	// 2. Set name and summary
	summary := plan.Summary.ValueString()
	visible := plan.Visible.ValueInt64()
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		visible = 1
	}

	err = r.client.EditSection(section.ID, plan.Name.ValueString(), summary, visible)
	if err != nil {
		resp.Diagnostics.AddError("Error editing section", err.Error())
		return
	}

	plan.ID = types.Int64Value(section.ID)
	plan.Section = types.Int64Value(section.Section)
	plan.Visible = types.Int64Value(visible)
	if plan.Summary.IsNull() || plan.Summary.IsUnknown() {
		plan.Summary = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *courseSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state courseSectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	section, err := r.client.GetSection(state.CourseID.ValueInt64(), state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading section", err.Error())
		return
	}

	state.Name = types.StringValue(section.Name)
	state.Summary = types.StringValue(section.Summary)
	state.Section = types.Int64Value(section.Section)
	state.Visible = types.Int64Value(section.Visible)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *courseSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan courseSectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	visible := plan.Visible.ValueInt64()
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		visible = 1
	}

	err := r.client.EditSection(plan.ID.ValueInt64(), plan.Name.ValueString(), plan.Summary.ValueString(), visible)
	if err != nil {
		resp.Diagnostics.AddError("Error updating section", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *courseSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state courseSectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSection(state.CourseID.ValueInt64(), state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting section", err.Error())
		return
	}
}
