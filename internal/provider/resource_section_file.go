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
			"Unexpected provider type",
			fmt.Sprintf("Expected *MoodleClient, got: %T. Please report this error.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *sectionFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Uploads a local file to Moodle and links it as a resource to a course section.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The Course Module ID (cmID) of the created resource module in Moodle.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course to which the file is added.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"section_num": schema.Int64Attribute{
				Required:    true,
				Description: "The section number (position in the course) to which the file is added.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"file_path": schema.StringAttribute{
				Required:    true,
				Description: "Relative or absolute path to the file to be uploaded. Relative paths are resolved relative to the working directory.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Display name of the file in Moodle. If not specified, the filename is used.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"visible": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Visibility of the file (1 = visible, 0 = hidden). Default: 1.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"file_hash": schema.StringAttribute{
				Optional:    true,
				Description: "MD5 hash of the file (e.g. filemd5(\"path/to/file\")). Changes force a re-upload.",
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

	// Filename as fallback for display_name
	filePath := plan.FilePath.ValueString()
	displayName := plan.DisplayName.ValueString()
	if plan.DisplayName.IsNull() || plan.DisplayName.IsUnknown() || displayName == "" {
		displayName = filepath.Base(filePath)
	}

	visible := plan.Visible.ValueInt64()
	if plan.Visible.IsNull() || plan.Visible.IsUnknown() {
		visible = 1
	}

	// 1. Upload file
	itemID, filename, err := r.client.UploadFile(filePath)
	if err != nil {
		resp.Diagnostics.AddError("Error uploading file", err.Error())
		return
	}

	// display_name: prefer user input, fallback to uploaded filename
	if plan.DisplayName.IsNull() || plan.DisplayName.IsUnknown() || plan.DisplayName.ValueString() == "" {
		displayName = filename
	}

	// 2. Link file to section
	cmID, err := r.client.AddFileToSection(
		plan.CourseID.ValueInt64(),
		plan.SectionNum.ValueInt64(),
		itemID,
		displayName,
		visible,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error adding file to section", err.Error())
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
		resp.Diagnostics.AddError("Error reading course module", err.Error())
		return
	}

	// Module was deleted externally — remove state
	if module == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.DisplayName = types.StringValue(module.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sectionFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All mutable attributes have RequiresReplace — Update never called.
	// Sets state anyway to keep Terraform consistent.
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
