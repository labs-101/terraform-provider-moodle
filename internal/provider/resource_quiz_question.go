package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"terraform-moodle-provider/internal/moodle"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// choiceObjectType describes the structure of a single "choice" nested block,
// used when (re)building the choice list from API responses.
var choiceObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"answer":   types.StringType,
		"grade":    types.Float64Type,
		"feedback": types.StringType,
	},
}

var (
	_ resource.Resource              = &quizQuestionResource{}
	_ resource.ResourceWithConfigure = &quizQuestionResource{}
)

func NewQuizQuestionResource() resource.Resource {
	return &quizQuestionResource{}
}

type quizQuestionResource struct {
	client *moodle.MoodleClient
}

type quizQuestionResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	CourseID     types.Int64  `tfsdk:"course_id"`
	QuizID       types.Int64  `tfsdk:"quiz_id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	QuestionText types.String `tfsdk:"question_text"`
	Choice       types.List   `tfsdk:"choice"`
	Slot         types.Int64  `tfsdk:"slot"`
	Page         types.Int64  `tfsdk:"page"`
}

type choiceModel struct {
	Answer   types.String  `tfsdk:"answer"`
	Grade    types.Float64 `tfsdk:"grade"`
	Feedback types.String  `tfsdk:"feedback"`
}

type choiceJSON struct {
	Answer   string  `json:"answer"`
	Grade    float64 `json:"grade"`
	Feedback string  `json:"feedback"`
}

type questionPayload struct {
	QuestionText string       `json:"questionText"`
	Choices      []choiceJSON `json:"choices"`
}

func buildChoicesJSON(ctx context.Context, questionText string, choiceList types.List) (string, error) {
	var choices []choiceModel
	diags := choiceList.ElementsAs(ctx, &choices, false)
	if diags.HasError() {
		return "", fmt.Errorf("error reading choices: %s", diags.Errors())
	}

	payload := questionPayload{
		QuestionText: questionText,
	}
	for _, c := range choices {
		payload.Choices = append(payload.Choices, choiceJSON{
			Answer:   c.Answer.ValueString(),
			Grade:    c.Grade.ValueFloat64(),
			Feedback: c.Feedback.ValueString(),
		})
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error encoding choices JSON: %w", err)
	}
	return string(data), nil
}

func (r *quizQuestionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quiz_question"
}

func (r *quizQuestionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *quizQuestionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adds a question to a Moodle quiz.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The question ID.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course containing the quiz.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"quiz_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the quiz to add the question to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the question.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The question type. Supported: 'multichoice'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"question_text": schema.StringAttribute{
				Required:    true,
				Description: "The HTML text of the question displayed to the student.",
			},
			"slot": schema.Int64Attribute{
				Computed:    true,
				Optional:    true,
				Description: "The slot position of the question in the quiz.",
			},
			"page": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The page number on which the question appears. Default: 0.",
			},
		},
		Blocks: map[string]schema.Block{
			"choice": schema.ListNestedBlock{
				Description: "Answer choice block. At least 2 choices are required for multichoice questions.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"answer": schema.StringAttribute{
							Required:    true,
							Description: "The answer text.",
						},
						"grade": schema.Float64Attribute{
							Required:    true,
							Description: "The grade for this choice (0.0 = incorrect, 1.0 = full credit). Supports partial credit values.",
						},
						"feedback": schema.StringAttribute{
							Required:    true,
							Description: "Feedback shown when this choice is selected.",
						},
					},
				},
			},
		},
	}
}

func (r *quizQuestionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan quizQuestionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	choicesJSON, err := buildChoicesJSON(ctx, plan.QuestionText.ValueString(), plan.Choice)
	if err != nil {
		resp.Diagnostics.AddError("Error building choices JSON", err.Error())
		return
	}

	page := plan.Page.ValueInt64()

	result, err := r.client.AddQuestionToQuiz(
		plan.CourseID.ValueInt64(),
		plan.Name.ValueString(),
		plan.QuizID.ValueInt64(),
		plan.Type.ValueString(),
		choicesJSON,
		page,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error adding question to quiz", err.Error())
		return
	}

	plan.ID = types.Int64Value(result.QuestionID)
	plan.Slot = types.Int64Value(result.Slot)
	plan.Page = types.Int64Value(result.Page)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *quizQuestionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state quizQuestionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	question, err := r.client.GetQuestion(
		state.CourseID.ValueInt64(),
		state.QuizID.ValueInt64(),
		state.ID.ValueInt64(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error reading question", err.Error())
		return
	}

	// The plugin resolves the question to its latest version, so refresh the ID
	// to keep subsequent updates/deletes pointed at the current version.
	state.ID = types.Int64Value(question.QuestionID)
	state.Name = types.StringValue(question.Name)
	state.Type = types.StringValue(question.Type)
	state.QuestionText = types.StringValue(question.QuestionText)
	state.Slot = types.Int64Value(question.Slot)
	state.Page = types.Int64Value(question.Page)

	choiceModels := make([]choiceModel, 0, len(question.Choices))
	for _, c := range question.Choices {
		choiceModels = append(choiceModels, choiceModel{
			Answer:   types.StringValue(c.Answer),
			Grade:    types.Float64Value(c.Grade),
			Feedback: types.StringValue(c.Feedback),
		})
	}

	choiceList, diags := types.ListValueFrom(ctx, choiceObjectType, choiceModels)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Choice = choiceList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *quizQuestionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan quizQuestionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state quizQuestionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	choicesJSON, err := buildChoicesJSON(ctx, plan.QuestionText.ValueString(), plan.Choice)
	if err != nil {
		resp.Diagnostics.AddError("Error building choices JSON", err.Error())
		return
	}

	slot := plan.Slot.ValueInt64()
	page := plan.Page.ValueInt64()

	result, err := r.client.UpdateQuestion(
		plan.CourseID.ValueInt64(),
		plan.QuizID.ValueInt64(),
		state.ID.ValueInt64(),
		plan.Name.ValueString(),
		plan.Type.ValueString(),
		choicesJSON,
		slot,
		page,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating question", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Slot = types.Int64Value(result.Slot)
	plan.Page = types.Int64Value(result.Page)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *quizQuestionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state quizQuestionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteQuestion(state.CourseID.ValueInt64(), state.QuizID.ValueInt64(), state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting question", err.Error())
	}
}
