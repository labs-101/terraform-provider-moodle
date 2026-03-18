package provider

import (
	"context"
	"fmt"
	"terraform-moodle-provider/internal/moodle"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	SectionNum               types.Int64  `tfsdk:"section_num"`
	Name                     types.String `tfsdk:"name"`
	Intro                    types.String `tfsdk:"intro"`
	DueDate                  types.String `tfsdk:"duedate"`
	AllowSubmissionsFromDate types.String `tfsdk:"allowsubmissionsfromdate"`
	MaxBytes                 types.Int64  `tfsdk:"maxbytes"`
	MaxFileSubmissions       types.Int64  `tfsdk:"maxfilesubmissions"`
	SubmissionTypes          types.String `tfsdk:"submissiontypes"`
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
			"Unerwarteter Provider-Typ",
			fmt.Sprintf("Erwartete *MoodleClient, bekam: %T. Bitte melde diesen Fehler.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *sectionAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Erstellt eine Assignment-Aktivität (Aufgabe) in einem Moodle-Kursabschnitt.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Die Course Module ID (cmID) der erstellten Aufgabe.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "Die ID des Kurses, zu dem die Aufgabe hinzugefügt wird.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"section_num": schema.Int64Attribute{
				Required:    true,
				Description: "Die Sektionsnummer (0-basiert), zu der die Aufgabe hinzugefügt wird.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Der Anzeigename der Aufgabe.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"intro": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Aufgabenbeschreibung (HTML wird unterstützt).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"duedate": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Abgabefrist im Format YYYY-MM-DD (z.B. 2026-06-30). Leer bedeutet keine Frist.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"allowsubmissionsfromdate": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Startdatum für Abgaben im Format YYYY-MM-DD. Leer bedeutet sofort.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"maxbytes": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximale Dateigröße in Bytes. 0 bedeutet unbegrenzt.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"maxfilesubmissions": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximale Anzahl hochladbarer Dateien pro Abgabe. Standard: 1.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"submissiontypes": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Abgabetypen als kommagetrennte Liste. Mögliche Werte: onlinetext, file. Standard: onlinetext.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

	// Defaults für optionale Felder
	intro := plan.Intro.ValueString()
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
		resp.Diagnostics.AddError("Ungültiges Abgabedatum", err.Error())
		return
	}
	allowFrom, err := parseDateToUnix(plan.AllowSubmissionsFromDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Ungültiges Startdatum für Abgaben", err.Error())
		return
	}

	cmID, err := r.client.AddAssignmentToSection(
		plan.CourseID.ValueInt64(),
		plan.SectionNum.ValueInt64(),
		plan.Name.ValueString(),
		intro,
		dueDate,
		allowFrom,
		maxBytes,
		maxFiles,
		submissionTypes,
	)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Erstellen der Aufgabe", err.Error())
		return
	}

	plan.ID = types.Int64Value(cmID)
	if plan.Intro.IsNull() || plan.Intro.IsUnknown() {
		plan.Intro = types.StringValue("")
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sectionAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sectionAssignmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	module, err := r.client.GetCourseModule(state.CourseID.ValueInt64(), state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Lesen der Aufgabe", err.Error())
		return
	}

	if module == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(module.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sectionAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Alle Attribute haben RequiresReplace — Update wird nie aufgerufen.
	var plan sectionAssignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sectionAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sectionAssignmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCourseModule(state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Fehler beim Löschen der Aufgabe", err.Error())
	}
}
