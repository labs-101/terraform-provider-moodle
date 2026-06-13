package provider

import (
	"context"
	"crypto/sha1"
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

// sha1HashFile computes the SHA-1 hex digest of the file at the given path.
// SHA-1 is used (not SHA-256) so the value matches Moodle's file content hash
// (mdl_files.contenthash), allowing remote drift to be detected by comparison.
func sha1HashFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// autoHashPlanModifier automatically computes the SHA-256 hash from file_path
// when file_hash is not set by the user, and triggers replacement on change.
type autoHashPlanModifier struct{}

func (m autoHashPlanModifier) Description(_ context.Context) string {
	return "Automatically computes the SHA-256 hash from file_path. Forces replacement when the hash changes."
}

func (m autoHashPlanModifier) MarkdownDescription(_ context.Context) string {
	return "Automatically computes the SHA-256 hash from `file_path`. Forces replacement when the hash changes."
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
		hash, err := sha1HashFile(filePath.ValueString())
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Could not auto-compute file hash",
				fmt.Sprintf("Failed to compute SHA-1 of %q: %s. Set file_hash explicitly.", filePath.ValueString(), err),
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

// autoFileSizePlanModifier tracks the byte size of the local file at file_path.
// It forces replacement when the size stored in state (refreshed from Moodle in
// Read) differs from the local file — i.e. a different file was uploaded to Moodle
// outside of Terraform, or the local file was swapped for one of a different size.
type autoFileSizePlanModifier struct{}

func (m autoFileSizePlanModifier) Description(_ context.Context) string {
	return "Tracks the local file size and forces replacement when the file in Moodle differs."
}

func (m autoFileSizePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Tracks the local file size and forces replacement when the file in Moodle differs."
}

func (m autoFileSizePlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	var filePath types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("file_path"), &filePath)...)
	if resp.Diagnostics.HasError() || filePath.IsNull() || filePath.IsUnknown() {
		return
	}

	info, err := os.Stat(filePath.ValueString())
	if err != nil {
		// Can't determine the local size — leave the planned value untouched.
		return
	}
	localSize := info.Size()

	resp.PlanValue = types.Int64Value(localSize)

	// Trigger replacement when the size in Moodle no longer matches the local file.
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		if req.StateValue.ValueInt64() != localSize {
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
	FileSize    types.Int64  `tfsdk:"file_size"`
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
					boolplanmodifier.RequiresReplace(),
				},
			},
			"file_hash": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SHA-1 content hash of the file (matches Moodle's file storage hash). If omitted, computed automatically from file_path. A mismatch — whether the local file changed or a different file was uploaded to Moodle — forces a re-upload.",
				PlanModifiers: []planmodifier.String{
					autoHashPlanModifier{},
				},
			},
			"file_size": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of the uploaded file in bytes, as reported by Moodle. If the size in Moodle no longer matches the local file (e.g. a different file was uploaded outside Terraform), the resource is replaced.",
				PlanModifiers: []planmodifier.Int64{
					autoFileSizePlanModifier{},
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

	itemID, filename, err := r.client.UploadFile(filePath)
	if err != nil {
		resp.Diagnostics.AddError("Error uploading file", err.Error())
		return
	}

	if plan.DisplayName.IsNull() || plan.DisplayName.IsUnknown() || plan.DisplayName.ValueString() == "" {
		displayName = filename
	}

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

	// Store hash in state compute as fallback if plan modifier could not set it.
	if plan.FileHash.IsNull() || plan.FileHash.IsUnknown() {
		if hash, err := sha1HashFile(filePath); err == nil {
			plan.FileHash = types.StringValue(hash)
		}
	}

	// Record the local file size; right after upload it matches Moodle's copy.
	if plan.FileSize.IsNull() || plan.FileSize.IsUnknown() {
		if info, err := os.Stat(filePath); err == nil {
			plan.FileSize = types.Int64Value(info.Size())
		} else {
			plan.FileSize = types.Int64Value(0)
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

	file, err := r.client.GetResourceFile(state.CourseID.ValueInt64(), state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading file resource", err.Error())
		return
	}

	// Activity was deleted externally — remove state
	if file == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Refresh tracked attributes so configuration/content drift is detected.
	state.DisplayName = types.StringValue(file.Name)
	state.Visible = types.BoolValue(file.Visible)

	// Refresh the content hash from Moodle (SHA-1). A mismatch with the local file's
	// hash makes the file_hash plan modifier force a replacement — this catches both
	// a changed local file and a different file uploaded to Moodle out-of-band.
	if file.ContentHash != "" {
		state.FileHash = types.StringValue(file.ContentHash)
	}

	// Keep the file size in sync as a secondary signal. Guard against transient
	// empty responses that would otherwise cause a spurious replacement.
	if file.FileSize > 0 {
		state.FileSize = types.Int64Value(file.FileSize)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sectionFileResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All attributes have RequiresReplace — Update is never called.
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
