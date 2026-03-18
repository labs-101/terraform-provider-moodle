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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &userEnrolmentResource{}
	_ resource.ResourceWithConfigure = &userEnrolmentResource{}
)

func NewUserEnrolmentResource() resource.Resource {
	return &userEnrolmentResource{}
}

type userEnrolmentResource struct {
	client *moodle.MoodleClient
}

type userEnrolmentResourceModel struct {
	ID        types.String `tfsdk:"id"`
	UserEmail types.String `tfsdk:"user_email"`
	CourseID  types.Int64  `tfsdk:"course_id"`
	RoleID    types.Int64  `tfsdk:"role_id"`
}

func (r *userEnrolmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_enrolment"
}

func (r *userEnrolmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userEnrolmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Moodle user enrolment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the enrolment (composite key).",
			},
			"user_email": schema.StringAttribute{
				Required:    true,
				Description: "The electronic mail address of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"role_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the role.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *userEnrolmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userEnrolmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUserByEmail(plan.UserEmail.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving user by email", err.Error())
		return
	}

	err = r.client.EnrolUser(user.ID, plan.CourseID.ValueInt64(), plan.RoleID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error enrolling user", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d-%d-%d", user.ID, plan.CourseID.ValueInt64(), plan.RoleID.ValueInt64()))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userEnrolmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userEnrolmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading enrolment for user: %s, course: %d, role: %d", state.UserEmail.ValueString(), state.CourseID.ValueInt64(), state.RoleID.ValueInt64()))

	users, err := r.client.GetEnrolledUsers(state.CourseID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading enrolled users", err.Error())
		return
	}

	found := false
	for _, user := range users {
		if user.Email == state.UserEmail.ValueString() {
			for _, role := range user.Roles {
				if role.RoleID == state.RoleID.ValueInt64() {
					found = true
					break
				}
			}
		}
		if found {
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userEnrolmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields ForceNew, so update is not needed (handled by plan modifiers)
}

func (r *userEnrolmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userEnrolmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUserByEmail(state.UserEmail.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving user by email", err.Error())
		return
	}

	err = r.client.UnenrolUser(user.ID, state.CourseID.ValueInt64(), state.RoleID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error unenrolling user", err.Error())
		return
	}
}
