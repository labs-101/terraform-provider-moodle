package provider

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"terraform-moodle-provider/internal/moodle"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &sectionFileResource{}
	_ resource.ResourceWithConfigure = &sectionFileResource{}
)

// md5HashFile computes the MD5 hex digest of the file at the given path.
func md5HashFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// autoHashPlanModifier automatically computes the MD5 hash from file_path
// when file_hash is not set by the user, and triggers replacement on change.
type autoHashPlanModifier struct{}

func (m autoHashPlanModifier) Description(_ context.Context) string {
	return "Automatically computes the MD5 hash from file_path. Forces replacement when the hash changes."
}

func (m autoHashPlanModifier) MarkdownDescription(_ context.Context) string {
	return "Automatically computes the MD5 hash from `file_path`. Forces replacement when the hash changes."
}

func (m autoHashPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var computedHash string

	if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
		// User provided an explicit hash — use it as-is.
		computedHash = req.ConfigValue.ValueString()
	} else {
		// Auto-compute from file_path.
		var filePath types.String
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("file_path"), &filePath)...)
		if resp.Diagnostics.HasError() || filePath.IsNull() || filePath.IsUnknown() {
			return
		}
		hash, err := md5HashFile(filePath.ValueString())
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Could not auto-compute file hash",
				fmt.Sprintf("Failed to compute MD5 of %q: %s. Set file_hash explicitly.", filePath.ValueString(), err),
			)
			return
		}
		computedHash = hash
	}

	resp.PlanValue = types.StringValue(computedHash)

	// Trigger replacement when hash changed from stored state.
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		if req.StateValue.ValueString() != computedHash {
			resp.RequiresReplace = true
		}
	}
}

func NewSectionFileResource() resource.Resource {
	return &sectionFileResource{}
}

type sectionFileResource struct {
	client *moodle.MoodleClient
}

type sectionFileResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	CourseID    types.Int64  `tfsdk:"course_id"`
	Section     types.Int64  `tfsdk:"section"`
	FilePath    types.String `tfsdk:"file_path"`
	DisplayName types.String `tfsdk:"display_name"`
	Visible     types.Bool   `tfsdk:"visible"`
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
			"section": schema.Int64Attribute{
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
			"visible": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the file is visible to students. Default: true.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"file_hash": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "MD5 hash of the file. If omitted, computed automatically from file_path. Changes force a re-upload.",
				PlanModifiers: []planmodifier.String{
					autoHashPlanModifier{},
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

	visible := true
	if !plan.Visible.IsNull() && !plan.Visible.IsUnknown() {
		visible = plan.Visible.ValueBool()
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
		plan.Section.ValueInt64(),
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
	plan.Visible = types.BoolValue(visible)

	// Store hash in state; compute as fallback if plan modifier could not set it.
	if plan.FileHash.IsNull() || plan.FileHash.IsUnknown() {
		if hash, err := md5HashFile(filePath); err == nil {
			plan.FileHash = types.StringValue(hash)
		}
	}

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
