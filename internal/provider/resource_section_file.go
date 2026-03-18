package provider

import (
	"context"
	"fmt"
	"path/filepath"
	"terraform-moodle-provider/internal/moodle"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &sectionFileResource{}
	_ resource.ResourceWithConfigure = &sectionFileResource{}
)

func NewSectionFileResource() resource.Resource {
	return &sectionFileResource{}
}

type sectionFileResource struct {
	client *moodle.MoodleClient
}

type sectionFileResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	CourseID    types.Int64  `tfsdk:"course_id"`
	SectionNum  types.Int64  `tfsdk:"section_num"`
	FilePath    types.String `tfsdk:"file_path"`
	DisplayName types.String `tfsdk:"display_name"`
	Visible     types.Int64  `tfsdk:"visible"`
	FileHash    types.String `tfsdk:"file_hash"`
}

func (r *sectionFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_section_file"
}

func (r *sectionFileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sectionFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lädt eine lokale Datei zu Moodle hoch und verknüpft sie als Ressource mit einer Kurs-Sektion.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Die Course Module ID (cmID) des erstellten Ressource-Moduls in Moodle.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "Die ID des Kurses, zu dem die Datei hinzugefügt wird.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"section_num": schema.Int64Attribute{
				Required:    true,
				Description: "Die Sektionsnummer (Position im Kurs), zu der die Datei hinzugefügt wird.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"file_path": schema.StringAttribute{
				Required:    true,
				Description: "Relativer oder absoluter Pfad zur hochzuladenden Datei. Relative Pfade werden relativ zum Arbeitsverzeichnis aufgelöst.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Anzeigename der Datei in Moodle. Wenn nicht angegeben, wird der Dateiname verwendet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"visible": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Sichtbarkeit der Datei (1 = sichtbar, 0 = verborgen). Standard: 1.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"file_hash": schema.StringAttribute{
				Optional:    true,
				Description: "MD5-Hash der Datei (z.B. filemd5(\"pfad/zur/datei\")). Änderungen erzwingen einen erneuten Upload.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *sectionFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sectionFileResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Dateiname als Fallback für display_name
	filePath := plan.FilePath.ValueString()
	displayName := plan.DisplayName.ValueString()
	if plan.DisplayName.IsNull() || plan.DisplayName.IsUnknown() || displayName == "" {
		displayName = filepath.Base(filePath)
	}

	visible := plan.Visible.ValueInt64()
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		visible = 1
	}

	// 1. Datei hochladen
	itemID, filename, err := r.client.UploadFile(filePath)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Hochladen der Datei", err.Error())
		return
	}

	// display_name: bevorzuge Benutzereingabe, Fallback auf hochgeladenen Dateinamen
	if plan.DisplayName.IsNull() || plan.DisplayName.IsUnknown() || plan.DisplayName.ValueString() == "" {
		displayName = filename
	}

	// 2. Datei mit Sektion verknüpfen
	cmID, err := r.client.AddFileToSection(
		plan.CourseID.ValueInt64(),
		plan.SectionNum.ValueInt64(),
		itemID,
		displayName,
		visible,
	)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Verknüpfen der Datei mit der Sektion", err.Error())
		return
	}

	plan.ID = types.Int64Value(cmID)
	plan.DisplayName = types.StringValue(displayName)
	plan.Visible = types.Int64Value(visible)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sectionFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sectionFileResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	module, err := r.client.GetCourseModule(state.CourseID.ValueInt64(), state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Lesen des Kurs-Moduls", err.Error())
		return
	}

	// Modul wurde extern gelöscht — State entfernen
	if module == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.DisplayName = types.StringValue(module.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sectionFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Alle veränderlichen Attribute haben RequiresReplace — Update wird nie aufgerufen.
	// Trotzdem State setzen, damit Terraform konsistent bleibt.
	var plan sectionFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sectionFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sectionFileResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCourseModule(state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Fehler beim Löschen des Kurs-Moduls", err.Error())
	}
}
