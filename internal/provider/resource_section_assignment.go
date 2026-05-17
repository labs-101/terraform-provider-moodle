package provider

import (
	"context"
	"fmt"
	"terraform-moodle-provider/internal/moodle"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &sectionAssignmentResource{}
	_ resource.ResourceWithConfigure = &sectionAssignmentResource{}
)

func NewSectionAssignmentResource() resource.Resource {
	return &sectionAssignmentResource{}
}

type sectionAssignmentResource struct {
	client *moodle.MoodleClient
}

type sectionAssignmentResourceModel struct {
	ID                       types.Int64  `tfsdk:"id"`
	CourseID                 types.Int64  `tfsdk:"course_id"`
	Section                  types.Int64  `tfsdk:"section"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	DueDate                  types.String `tfsdk:"duedate"`
	AllowSubmissionsFromDate types.String `tfsdk:"allowsubmissionsfromdate"`
	MaxBytes                 types.Int64  `tfsdk:"maxbytes"`
	MaxFileSubmissions       types.Int64  `tfsdk:"maxfilesubmissions"`
	SubmissionTypes          types.String `tfsdk:"submissiontypes"`
	Visible                  types.Bool   `tfsdk:"visible"`
}

func (r *sectionAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_section_assignment"
}

func (r *sectionAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sectionAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an assignment activity in a Moodle course section.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The Course Module ID (cmID) of the created assignment.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course to which the assignment is added.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"section": schema.Int64Attribute{
				Required:    true,
				Description: "The section number (0-based) to which the assignment is added.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the assignment.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Assignment description (HTML is supported).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"duedate": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Due date in format YYYY-MM-DD (e.g. 2026-06-30). Empty means no due date.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allowsubmissionsfromdate": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Start date for submissions in format YYYY-MM-DD. Empty means immediately.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"maxbytes": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum file size in bytes. 0 means unlimited.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"maxfilesubmissions": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of file submissions. Default: 1.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"submissiontypes": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Submission types as comma-separated list. Possible values: onlinetext, file. Default: onlinetext.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"visible": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the assignment is visible to students. Default: true.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *sectionAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sectionAssignmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	description := plan.Description.ValueString()
	maxBytes := plan.MaxBytes.ValueInt64()
	maxFiles := plan.MaxFileSubmissions.ValueInt64()
	if plan.MaxFileSubmissions.IsNull() || plan.MaxFileSubmissions.IsUnknown() {
		maxFiles = 1
	}
	submissionTypes := plan.SubmissionTypes.ValueString()
	if plan.SubmissionTypes.IsNull() || plan.SubmissionTypes.IsUnknown() || submissionTypes == "" {
		submissionTypes = "onlinetext"
	}

	dueDate, err := parseDateToUnix(plan.DueDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid due date", err.Error())
		return
	}
	allowFrom, err := parseDateToUnix(plan.AllowSubmissionsFromDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid start date for submissions", err.Error())
		return
	}

	visible := true
	if !plan.Visible.IsNull() && !plan.Visible.IsUnknown() {
		visible = plan.Visible.ValueBool()
	}

	cmID, err := r.client.AddAssignmentToSection(
		plan.CourseID.ValueInt64(),
		plan.Section.ValueInt64(),
		plan.Name.ValueString(),
		description,
		dueDate,
		allowFrom,
		maxBytes,
		maxFiles,
		submissionTypes,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating assignment", err.Error())
		return
	}

	// AddAssignmentToSection doesn't accept visible — hide immediately if needed.
	if !visible {
		if err := r.client.UpdateAssignment(
			plan.CourseID.ValueInt64(),
			cmID,
			plan.Name.ValueString(),
			description,
			dueDate,
			allowFrom,
			maxBytes,
			maxFiles,
			submissionTypes,
			false,
		); err != nil {
			resp.Diagnostics.AddError("Error setting assignment visibility", err.Error())
			return
		}
	}

	plan.ID = types.Int64Value(cmID)
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		plan.Description = types.StringValue("")
	}
	if plan.DueDate.IsNull() || plan.DueDate.IsUnknown() {
		plan.DueDate = types.StringValue("")
	}
	if plan.AllowSubmissionsFromDate.IsNull() || plan.AllowSubmissionsFromDate.IsUnknown() {
		plan.AllowSubmissionsFromDate = types.StringValue("")
	}
	if plan.MaxBytes.IsNull() || plan.MaxBytes.IsUnknown() {
		plan.MaxBytes = types.Int64Value(0)
	}
	if plan.MaxFileSubmissions.IsNull() || plan.MaxFileSubmissions.IsUnknown() {
		plan.MaxFileSubmissions = types.Int64Value(maxFiles)
	}
	if plan.SubmissionTypes.IsNull() || plan.SubmissionTypes.IsUnknown() {
		plan.SubmissionTypes = types.StringValue(submissionTypes)
	}
	plan.Visible = types.BoolValue(visible)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sectionAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sectionAssignmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assignment, err := r.client.GetAssignment(state.CourseID.ValueInt64(), state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading assignment", err.Error())
		return
	}

	state.Name = types.StringValue(assignment.Name)
	state.Description = types.StringValue(assignment.Intro)
	state.DueDate = types.StringValue(unixToDate(assignment.DueDate))
	state.AllowSubmissionsFromDate = types.StringValue(unixToDate(assignment.AllowSubmissionsFromDate))
	state.MaxBytes = types.Int64Value(assignment.MaxBytes)
	state.MaxFileSubmissions = types.Int64Value(assignment.MaxFileSubmissions)
	state.SubmissionTypes = types.StringValue(assignment.SubmissionTypes)
	state.Visible = types.BoolValue(assignment.Visible)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sectionAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sectionAssignmentResourceModel
	var state sectionAssignmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	description := plan.Description.ValueString()
	maxBytes := plan.MaxBytes.ValueInt64()
	maxFiles := plan.MaxFileSubmissions.ValueInt64()
	if plan.MaxFileSubmissions.IsNull() || plan.MaxFileSubmissions.IsUnknown() {
		maxFiles = 1
	}
	submissionTypes := plan.SubmissionTypes.ValueString()
	if plan.SubmissionTypes.IsNull() || plan.SubmissionTypes.IsUnknown() || submissionTypes == "" {
		submissionTypes = "onlinetext"
	}

	dueDate, err := parseDateToUnix(plan.DueDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid due date", err.Error())
		return
	}
	allowFrom, err := parseDateToUnix(plan.AllowSubmissionsFromDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid start date for submissions", err.Error())
		return
	}

	visible := true
	if !plan.Visible.IsNull() && !plan.Visible.IsUnknown() {
		visible = plan.Visible.ValueBool()
	}

	err = r.client.UpdateAssignment(
		state.CourseID.ValueInt64(),
		state.ID.ValueInt64(),
		plan.Name.ValueString(),
		description,
		dueDate,
		allowFrom,
		maxBytes,
		maxFiles,
		submissionTypes,
		visible,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating assignment", err.Error())
		return
	}

	plan.ID = state.ID
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		plan.Description = types.StringValue("")
	}
	if plan.DueDate.IsNull() || plan.DueDate.IsUnknown() {
		plan.DueDate = types.StringValue("")
	}
	if plan.AllowSubmissionsFromDate.IsNull() || plan.AllowSubmissionsFromDate.IsUnknown() {
		plan.AllowSubmissionsFromDate = types.StringValue("")
	}
	if plan.MaxBytes.IsNull() || plan.MaxBytes.IsUnknown() {
		plan.MaxBytes = types.Int64Value(0)
	}
	if plan.MaxFileSubmissions.IsNull() || plan.MaxFileSubmissions.IsUnknown() {
		plan.MaxFileSubmissions = types.Int64Value(maxFiles)
	}
	if plan.SubmissionTypes.IsNull() || plan.SubmissionTypes.IsUnknown() {
		plan.SubmissionTypes = types.StringValue(submissionTypes)
	}
	plan.Visible = types.BoolValue(visible)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sectionAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sectionAssignmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCourseModule(state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting assignment", err.Error())
	}
}
