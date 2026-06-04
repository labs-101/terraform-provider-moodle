package provider

import (
	"context"
	"fmt"
	"slices"
	"terraform-moodle-provider/internal/moodle"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &groupMemberResource{}
	_ resource.ResourceWithConfigure = &groupMemberResource{}
)

func NewGroupMemberResource() resource.Resource {
	return &groupMemberResource{}
}

type groupMemberResource struct {
	client *moodle.MoodleClient
}

type groupMemberResourceModel struct {
	ID       types.String `tfsdk:"id"`
	CourseID types.Int64  `tfsdk:"course_id"`
	GroupID  types.Int64  `tfsdk:"group_id"`
	UserID   types.Int64  `tfsdk:"user_id"`
}

func (r *groupMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_member"
}

func (r *groupMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *groupMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adds a user to a Moodle group. The user must already be enrolled in the course.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite ID of the group membership (groupid-userid).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course the group belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"group_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the user to add to the group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *groupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupMemberResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddMemberToGroup(plan.GroupID.ValueInt64(), plan.UserID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error adding member to group", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d-%d", plan.GroupID.ValueInt64(), plan.UserID.ValueInt64()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	members, err := r.client.GetGroupMembers(state.CourseID.ValueInt64(), state.GroupID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading group members", err.Error())
		return
	}

	if !slices.Contains(members, state.UserID.ValueInt64()) {
		// The user is no longer a member of the group — drop it from state so it gets recreated.
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *groupMemberResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Both group_id and user_id have RequiresReplace — Update is never called.
}

func (r *groupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RemoveMemberFromGroup(state.GroupID.ValueInt64(), state.UserID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error removing member from group", err.Error())
	}
}
