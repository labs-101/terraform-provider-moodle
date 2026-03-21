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
	_ resource.Resource              = &sectionChoiceResource{}
	_ resource.ResourceWithConfigure = &sectionChoiceResource{}
)

func NewSectionChoiceResource() resource.Resource {
	return &sectionChoiceResource{}
}

type sectionChoiceResource struct {
	client *moodle.MoodleClient
}

type sectionChoiceResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	CourseID      types.Int64  `tfsdk:"course_id"`
	SectionNum    types.Int64  `tfsdk:"section_num"`
	Name          types.String `tfsdk:"name"`
	Intro         types.String `tfsdk:"intro"`
	Options       types.List   `tfsdk:"options"`
	AllowMultiple types.Bool   `tfsdk:"allow_multiple"`
}

func (r *sectionChoiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_section_choice"
}

func (r *sectionChoiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sectionChoiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a Choice activity in a Moodle course section.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The Course Module ID (cmID) of the created Choice activity.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course to which the Choice is added.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"section_num": schema.Int64Attribute{
				Required:    true,
				Description: "The section number (0-based) to which the Choice is added.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the Choice activity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"intro": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description text of the Choice activity (HTML is supported).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"options": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of options (at least 2).",
				PlanModifiers: []planmodifier.List{
					listRequiresReplace{},
				},
			},
			"allow_multiple": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether multiple selection is allowed. Default: false.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *sectionChoiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sectionChoiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// extract options from types.List
	var options []string
	resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &options, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intro := plan.Intro.ValueString()
	allowMultiple := plan.AllowMultiple.ValueBool()

	cmID, err := r.client.AddChoiceToSection(
		plan.CourseID.ValueInt64(),
		plan.SectionNum.ValueInt64(),
		plan.Name.ValueString(),
		intro,
		options,
		allowMultiple,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Choice activity", err.Error())
		return
	}

	plan.ID = types.Int64Value(cmID)
	if plan.Intro.IsNull() || plan.Intro.IsUnknown() {
		plan.Intro = types.StringValue("")
	}
	if plan.AllowMultiple.IsNull() || plan.AllowMultiple.IsUnknown() {
		plan.AllowMultiple = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sectionChoiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sectionChoiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	module, err := r.client.GetCourseModule(state.CourseID.ValueInt64(), state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Choice activity", err.Error())
		return
	}

	if module == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(module.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sectionChoiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes have RequiresReplace — Update never called.
	var plan sectionChoiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sectionChoiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sectionChoiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCourseModule(state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Fehler beim Löschen der Choice-Aktivität", err.Error())
	}
}

// listRequiresReplace ist ein einfacher PlanModifier für Listen, der bei Änderungen einen Replace erzwingt.
type listRequiresReplace struct{}

func (m listRequiresReplace) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}

func (m listRequiresReplace) Description(ctx context.Context) string {
	return "Erzwingt eine Neuanlage wenn sich die Liste ändert."
}

func (m listRequiresReplace) MarkdownDescription(ctx context.Context) string {
	return "Erzwingt eine Neuanlage wenn sich die Liste ändert."
}
