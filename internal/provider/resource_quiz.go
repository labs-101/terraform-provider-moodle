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
	_ resource.Resource              = &quizResource{}
	_ resource.ResourceWithConfigure = &quizResource{}
)

func NewQuizResource() resource.Resource {
	return &quizResource{}
}

type quizResource struct {
	client *moodle.MoodleClient
}

type quizResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	CourseID         types.Int64  `tfsdk:"course_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Password         types.String `tfsdk:"password"`
	TimeOpen         types.String `tfsdk:"timeopen"`
	TimeClose        types.String `tfsdk:"timeclose"`
	TimeLimit        types.Int64  `tfsdk:"timelimit"`
	Attempts         types.Int64  `tfsdk:"attempts"`
	GradeMethod      types.Int64  `tfsdk:"grademethod"`
	QuestionsPerPage types.Int64  `tfsdk:"questionsperpage"`
	NavMethod        types.String `tfsdk:"navmethod"`
	Section          types.Int64  `tfsdk:"section"`
	Visible          types.Int64  `tfsdk:"visible"`
}

func (r *quizResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quiz"
}

func (r *quizResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *quizResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a quiz activity in a Moodle course.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The quiz instance ID.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course to which the quiz is added.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the quiz.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Quiz description (HTML is supported).",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "Password required to access the quiz. Empty means no password.",
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
			"timelimit": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Time limit in seconds. 0 means no time limit.",
			},
			"attempts": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of attempts. 0 means unlimited.",
			},
			"grademethod": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Grading method: 1=Highest, 2=Average, 3=First, 4=Last.",
			},
			"questionsperpage": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of questions per page. Default: 1.",
			},
			"navmethod": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Navigation method: 'free' or 'sequential'. Default: 'free'.",
			},
			"section": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The section number (0-based) to place the quiz in. Default: 0.",
			},
			"visible": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Visibility: 1=visible, 0=hidden. Default: 1.",
			},
		},
	}
}

func (r *quizResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan quizResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intro := plan.Description.ValueString()
	password := plan.Password.ValueString()
	navMethod := plan.NavMethod.ValueString()
	if plan.NavMethod.IsNull() || plan.NavMethod.IsUnknown() || navMethod == "" {
		navMethod = "free"
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

	timeLimit := plan.TimeLimit.ValueInt64()
	attempts := plan.Attempts.ValueInt64()
	gradeMethod := plan.GradeMethod.ValueInt64()
	if plan.GradeMethod.IsNull() || plan.GradeMethod.IsUnknown() {
		gradeMethod = 1
	}
	questionsPerPage := plan.QuestionsPerPage.ValueInt64()
	if plan.QuestionsPerPage.IsNull() || plan.QuestionsPerPage.IsUnknown() {
		questionsPerPage = 1
	}
	section := plan.Section.ValueInt64()
	visible := plan.Visible.ValueInt64()
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		visible = 1
	}

	quizID, err := r.client.CreateQuiz(
		plan.CourseID.ValueInt64(),
		plan.Name.ValueString(),
		intro,
		password,
		timeOpen,
		timeClose,
		timeLimit,
		attempts,
		gradeMethod,
		questionsPerPage,
		navMethod,
		section,
		visible,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating quiz", err.Error())
		return
	}

	plan.ID = types.Int64Value(quizID)
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		plan.Description = types.StringValue("")
	}
	if plan.Password.IsNull() || plan.Password.IsUnknown() {
		plan.Password = types.StringValue("")
	}
	if plan.TimeOpen.IsNull() || plan.TimeOpen.IsUnknown() {
		plan.TimeOpen = types.StringValue("")
	}
	if plan.TimeClose.IsNull() || plan.TimeClose.IsUnknown() {
		plan.TimeClose = types.StringValue("")
	}
	if plan.TimeLimit.IsNull() || plan.TimeLimit.IsUnknown() {
		plan.TimeLimit = types.Int64Value(0)
	}
	if plan.Attempts.IsNull() || plan.Attempts.IsUnknown() {
		plan.Attempts = types.Int64Value(0)
	}
	if plan.GradeMethod.IsNull() || plan.GradeMethod.IsUnknown() {
		plan.GradeMethod = types.Int64Value(gradeMethod)
	}
	if plan.QuestionsPerPage.IsNull() || plan.QuestionsPerPage.IsUnknown() {
		plan.QuestionsPerPage = types.Int64Value(questionsPerPage)
	}
	if plan.NavMethod.IsNull() || plan.NavMethod.IsUnknown() {
		plan.NavMethod = types.StringValue(navMethod)
	}
	if plan.Section.IsNull() || plan.Section.IsUnknown() {
		plan.Section = types.Int64Value(section)
	}
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		plan.Visible = types.Int64Value(visible)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *quizResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state quizResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *quizResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan quizResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state quizResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intro := plan.Description.ValueString()
	password := plan.Password.ValueString()
	navMethod := plan.NavMethod.ValueString()
	if plan.NavMethod.IsNull() || plan.NavMethod.IsUnknown() || navMethod == "" {
		navMethod = "free"
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

	timeLimit := plan.TimeLimit.ValueInt64()
	attempts := plan.Attempts.ValueInt64()
	gradeMethod := plan.GradeMethod.ValueInt64()
	if plan.GradeMethod.IsNull() || plan.GradeMethod.IsUnknown() {
		gradeMethod = 1
	}
	questionsPerPage := plan.QuestionsPerPage.ValueInt64()
	if plan.QuestionsPerPage.IsNull() || plan.QuestionsPerPage.IsUnknown() {
		questionsPerPage = 1
	}
	section := plan.Section.ValueInt64()
	visible := plan.Visible.ValueInt64()
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		visible = 1
	}

	err = r.client.UpdateQuiz(
		state.ID.ValueInt64(),
		plan.Name.ValueString(),
		intro,
		password,
		timeOpen,
		timeClose,
		timeLimit,
		attempts,
		gradeMethod,
		questionsPerPage,
		navMethod,
		section,
		visible,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating quiz", err.Error())
		return
	}

	plan.ID = state.ID
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		plan.Description = types.StringValue("")
	}
	if plan.Password.IsNull() || plan.Password.IsUnknown() {
		plan.Password = types.StringValue("")
	}
	if plan.TimeOpen.IsNull() || plan.TimeOpen.IsUnknown() {
		plan.TimeOpen = types.StringValue("")
	}
	if plan.TimeClose.IsNull() || plan.TimeClose.IsUnknown() {
		plan.TimeClose = types.StringValue("")
	}
	if plan.TimeLimit.IsNull() || plan.TimeLimit.IsUnknown() {
		plan.TimeLimit = types.Int64Value(0)
	}
	if plan.Attempts.IsNull() || plan.Attempts.IsUnknown() {
		plan.Attempts = types.Int64Value(0)
	}
	if plan.GradeMethod.IsNull() || plan.GradeMethod.IsUnknown() {
		plan.GradeMethod = types.Int64Value(gradeMethod)
	}
	if plan.QuestionsPerPage.IsNull() || plan.QuestionsPerPage.IsUnknown() {
		plan.QuestionsPerPage = types.Int64Value(questionsPerPage)
	}
	if plan.NavMethod.IsNull() || plan.NavMethod.IsUnknown() {
		plan.NavMethod = types.StringValue(navMethod)
	}
	if plan.Section.IsNull() || plan.Section.IsUnknown() {
		plan.Section = types.Int64Value(section)
	}
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		plan.Visible = types.Int64Value(visible)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *quizResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state quizResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteQuiz(state.CourseID.ValueInt64(), state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting quiz", err.Error())
	}
}
