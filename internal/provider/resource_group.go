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
	_ resource.Resource              = &groupResource{}
	_ resource.ResourceWithConfigure = &groupResource{}
)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	client *moodle.MoodleClient
}

type groupResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	CourseID      types.Int64  `tfsdk:"course_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Enrolmentkey  types.String `tfsdk:"enrolmentkey"`
	Visibility    types.Int64  `tfsdk:"visibility"`
	Participation types.Int64  `tfsdk:"participation"`
	IDNumber      types.String `tfsdk:"idnumber"`
}

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a group in a Moodle course.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the created group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course in which the group is created.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the group.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the group (HTML supported).",
			},
			"enrolmentkey": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "Enrolment key for the group. Empty means no key.",
			},
			"visibility": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Visibility: 0=all members, 1=not in group, 2=only members, 3=none. Default: 0.",
			},
			"participation": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the group is a participation group. 1=yes, 0=no. Default: 1.",
			},
			"idnumber": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "An optional ID number for the group.",
			},
		},
	}
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	visibility := plan.Visibility.ValueInt64()
	participation := plan.Participation.ValueInt64()
	if plan.Participation.IsNull() || plan.Participation.IsUnknown() {
		participation = 1
	}

	groupID, err := r.client.CreateGroup(
		plan.CourseID.ValueInt64(),
		plan.Name.ValueString(),
		plan.Description.ValueString(),
		plan.Enrolmentkey.ValueString(),
		visibility,
		participation,
		plan.IDNumber.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating group", err.Error())
		return
	}

	plan.ID = types.Int64Value(groupID)
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		plan.Description = types.StringValue("")
	}
	if plan.Enrolmentkey.IsNull() || plan.Enrolmentkey.IsUnknown() {
		plan.Enrolmentkey = types.StringValue("")
	}
	if plan.Visibility.IsNull() || plan.Visibility.IsUnknown() {
		plan.Visibility = types.Int64Value(0)
	}
	if plan.Participation.IsNull() || plan.Participation.IsUnknown() {
		plan.Participation = types.Int64Value(participation)
	}
	if plan.IDNumber.IsNull() || plan.IDNumber.IsUnknown() {
		plan.IDNumber = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No dedicated read endpoint in the custom API — keep existing state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupResourceModel
	var state groupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	participation := plan.Participation.ValueInt64()
	if plan.Participation.IsNull() || plan.Participation.IsUnknown() {
		participation = 1
	}

	err := r.client.UpdateGroup(
		state.ID.ValueInt64(),
		plan.Name.ValueString(),
		plan.Description.ValueString(),
		plan.Enrolmentkey.ValueString(),
		plan.Visibility.ValueInt64(),
		participation,
		plan.IDNumber.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating group", err.Error())
		return
	}

	plan.ID = state.ID
	if plan.Description.IsNull() || plan.Description.IsUnknown() {
		plan.Description = types.StringValue("")
	}
	if plan.Enrolmentkey.IsNull() || plan.Enrolmentkey.IsUnknown() {
		plan.Enrolmentkey = types.StringValue("")
	}
	if plan.Visibility.IsNull() || plan.Visibility.IsUnknown() {
		plan.Visibility = types.Int64Value(0)
	}
	if plan.Participation.IsNull() || plan.Participation.IsUnknown() {
		plan.Participation = types.Int64Value(participation)
	}
	if plan.IDNumber.IsNull() || plan.IDNumber.IsUnknown() {
		plan.IDNumber = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteGroup(state.CourseID.ValueInt64(), state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting group", err.Error())
	}
}
