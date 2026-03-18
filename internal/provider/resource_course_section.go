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
			"Unerwarteter Provider-Typ",
			fmt.Sprintf("Erwartete *MoodleClient, bekam: %T. Bitte melde diesen Fehler.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *courseSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Verwaltet eine Sektion (Abschnitt) in einem Moodle-Kurs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Die interne Datenbank-ID der Moodle-Sektion.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "Die ID des Kurses, zu dem diese Sektion gehört.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Der Anzeigename der Sektion.",
			},
			"summary": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Die Zusammenfassung/Beschreibung der Sektion (HTML wird unterstützt).",
			},
			"section": schema.Int64Attribute{
				Computed:    true,
				Description: "Die Sektionsnummer (Position) innerhalb des Kurses, wird von Moodle vergeben.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"visible": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Sichtbarkeit der Sektion (1 = sichtbar, 0 = verborgen). Standard: 1.",
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

	// 1. Neue Sektion zum Kurs hinzufügen (Moodle hängt sie ans Ende)
	section, err := r.client.CreateSection(plan.CourseID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Erstellen der Sektion", err.Error())
		return
	}

	// 2. Name und Zusammenfassung setzen
	summary := plan.Summary.ValueString()
	visible := plan.Visible.ValueInt64()
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		visible = 1
	}

	err = r.client.EditSection(section.ID, plan.Name.ValueString(), summary, visible)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Bearbeiten der Sektion", err.Error())
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
		resp.Diagnostics.AddError("Fehler beim Lesen der Sektion", err.Error())
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
		resp.Diagnostics.AddError("Fehler beim Aktualisieren der Sektion", err.Error())
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
		resp.Diagnostics.AddError("Fehler beim Löschen der Sektion", err.Error())
		return
	}
}
