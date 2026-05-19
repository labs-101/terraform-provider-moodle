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
	Section       types.Int64  `tfsdk:"section"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
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
			"section": schema.Int64Attribute{
				Required:    true,
				Description: "The section number (0-based) to which the Choice is added.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the Choice activity.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description text of the Choice activity (HTML is supported).",
			},
			"options": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of options (at least 2).",
			},
			"allow_multiple": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether multiple selection is allowed. Default: false.",
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

	var options []string
	resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &options, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	description := plan.Description.ValueString()
	allowMultiple := plan.AllowMultiple.ValueBool()

	cmID, err := r.client.AddChoiceToSection(
		plan.CourseID.ValueInt64(),
		plan.Section.ValueInt64(),
		plan.Name.ValueString(),
		description,
		options,
		allowMultiple,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Choice activity", err.Error())
		return
	}

	plan.ID = types.Int64Value(cmID)
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		plan.Description = types.StringValue("")
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

	choice, err := r.client.GetChoice(state.CourseID.ValueInt64(), state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Choice activity", err.Error())
		return
	}

	if choice == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(choice.Name)
	state.Description = types.StringValue(choice.Intro)
	state.AllowMultiple = types.BoolValue(choice.AllowMultiple)

	opts, diags := types.ListValueFrom(ctx, types.StringType, choice.Options)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Options = opts

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sectionChoiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sectionChoiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var options []string
	resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &options, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateChoice(
		plan.CourseID.ValueInt64(),
		plan.ID.ValueInt64(),
		plan.Name.ValueString(),
		plan.Description.ValueString(),
		options,
		plan.AllowMultiple.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Choice activity", err.Error())
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
