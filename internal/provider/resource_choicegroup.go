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
	_ resource.Resource              = &choicegroupResource{}
	_ resource.ResourceWithConfigure = &choicegroupResource{}
)

func NewChoicegroupResource() resource.Resource {
	return &choicegroupResource{}
}

type choicegroupResource struct {
	client *moodle.MoodleClient
}

type choicegroupResourceModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	CourseID            types.Int64  `tfsdk:"course_id"`
	SectionNum          types.Int64  `tfsdk:"section_num"`
	Name                types.String `tfsdk:"name"`
	Intro               types.String `tfsdk:"intro"`
	GroupIDs            types.List   `tfsdk:"group_ids"`
	MultipleEnrollments types.Int64  `tfsdk:"multipleenrollmentspossible"`
	ShowResults         types.Int64  `tfsdk:"showresults"`
	AllowUpdate         types.Int64  `tfsdk:"allowupdate"`
	TimeOpen            types.String `tfsdk:"timeopen"`
	TimeClose           types.String `tfsdk:"timeclose"`
	Visible             types.Int64  `tfsdk:"visible"`
	PreviousElementID   types.Int64  `tfsdk:"previous_element_id"`
}

func (r *choicegroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_choicegroup"
}

func (r *choicegroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *choicegroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a Group Choice (choicegroup) activity in a Moodle course section.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The Course Module ID (cmID) of the created choicegroup activity.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"section_num": schema.Int64Attribute{
				Required:    true,
				Description: "The section number (0-based) to place the activity in. Can be changed via update.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the Group Choice activity.",
			},
			"intro": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description text (HTML supported).",
			},
			"group_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
				Description: "List of group IDs that participants can choose from (at least 1).",
			},
			"multipleenrollmentspossible": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Allow enrolment in multiple groups. 1=yes, 0=no. Default: 0.",
			},
			"showresults": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "When to show results: 0=never, 1=after answer, 2=after close, 3=always. Default: 0.",
			},
			"allowupdate": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Allow participants to change their choice. 1=yes, 0=no. Default: 0.",
			},
			"timeopen": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Opening date in format YYYY-MM-DD. Empty means immediately available.",
			},
			"timeclose": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Closing date in format YYYY-MM-DD. Empty means no closing date.",
			},
			"visible": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Visibility: 1=visible, 0=hidden. Default: 1.",
			},
			"previous_element_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The cmID of the element this activity should be placed after. 0 means no specific ordering.",
			},
		},
	}
}

func (r *choicegroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan choicegroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var groupIDs []int64
	resp.Diagnostics.Append(plan.GroupIDs.ElementsAs(ctx, &groupIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeOpen, err := parseDateToUnix(plan.TimeOpen.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeopen date", err.Error())
		return
	}
	timeClose, err := parseDateToUnix(plan.TimeClose.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeclose date", err.Error())
		return
	}

	visible := plan.Visible.ValueInt64()
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		visible = 1
	}

	cmID, err := r.client.CreateChoicegroup(
		plan.CourseID.ValueInt64(),
		plan.SectionNum.ValueInt64(),
		plan.Name.ValueString(),
		plan.Intro.ValueString(),
		groupIDs,
		plan.MultipleEnrollments.ValueInt64(),
		plan.ShowResults.ValueInt64(),
		plan.AllowUpdate.ValueInt64(),
		timeOpen,
		timeClose,
		visible,
		plan.PreviousElementID.ValueInt64(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating choicegroup", err.Error())
		return
	}

	plan.ID = types.Int64Value(cmID)
	if plan.Intro.IsNull() || plan.Intro.IsUnknown() {
		plan.Intro = types.StringValue("")
	}
	if plan.MultipleEnrollments.IsNull() || plan.MultipleEnrollments.IsUnknown() {
		plan.MultipleEnrollments = types.Int64Value(0)
	}
	if plan.ShowResults.IsNull() || plan.ShowResults.IsUnknown() {
		plan.ShowResults = types.Int64Value(0)
	}
	if plan.AllowUpdate.IsNull() || plan.AllowUpdate.IsUnknown() {
		plan.AllowUpdate = types.Int64Value(0)
	}
	if plan.TimeOpen.IsNull() || plan.TimeOpen.IsUnknown() {
		plan.TimeOpen = types.StringValue("")
	}
	if plan.TimeClose.IsNull() || plan.TimeClose.IsUnknown() {
		plan.TimeClose = types.StringValue("")
	}
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		plan.Visible = types.Int64Value(visible)
	}
	if plan.PreviousElementID.IsNull() || plan.PreviousElementID.IsUnknown() {
		plan.PreviousElementID = types.Int64Value(0)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *choicegroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state choicegroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No dedicated read endpoint in the custom API — keep existing state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *choicegroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan choicegroupResourceModel
	var state choicegroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var groupIDs []int64
	resp.Diagnostics.Append(plan.GroupIDs.ElementsAs(ctx, &groupIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeOpen, err := parseDateToUnix(plan.TimeOpen.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeopen date", err.Error())
		return
	}
	timeClose, err := parseDateToUnix(plan.TimeClose.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeclose date", err.Error())
		return
	}

	visible := plan.Visible.ValueInt64()
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		visible = 1
	}

	// Pass section_num as section to allow repositioning.
	section := plan.SectionNum.ValueInt64()

	err = r.client.UpdateChoicegroup(
		state.CourseID.ValueInt64(),
		state.ID.ValueInt64(),
		plan.Name.ValueString(),
		plan.Intro.ValueString(),
		groupIDs,
		plan.MultipleEnrollments.ValueInt64(),
		plan.ShowResults.ValueInt64(),
		plan.AllowUpdate.ValueInt64(),
		timeOpen,
		timeClose,
		visible,
		section,
		plan.PreviousElementID.ValueInt64(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating choicegroup", err.Error())
		return
	}

	plan.ID = state.ID
	if plan.Intro.IsNull() || plan.Intro.IsUnknown() {
		plan.Intro = types.StringValue("")
	}
	if plan.MultipleEnrollments.IsNull() || plan.MultipleEnrollments.IsUnknown() {
		plan.MultipleEnrollments = types.Int64Value(0)
	}
	if plan.ShowResults.IsNull() || plan.ShowResults.IsUnknown() {
		plan.ShowResults = types.Int64Value(0)
	}
	if plan.AllowUpdate.IsNull() || plan.AllowUpdate.IsUnknown() {
		plan.AllowUpdate = types.Int64Value(0)
	}
	if plan.TimeOpen.IsNull() || plan.TimeOpen.IsUnknown() {
		plan.TimeOpen = types.StringValue("")
	}
	if plan.TimeClose.IsNull() || plan.TimeClose.IsUnknown() {
		plan.TimeClose = types.StringValue("")
	}
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		plan.Visible = types.Int64Value(visible)
	}
	if plan.PreviousElementID.IsNull() || plan.PreviousElementID.IsUnknown() {
		plan.PreviousElementID = types.Int64Value(0)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *choicegroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state choicegroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteModule(state.CourseID.ValueInt64(), state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting choicegroup", err.Error())
	}
}
