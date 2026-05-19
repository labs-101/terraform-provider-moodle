package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"terraform-moodle-provider/internal/moodle"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *moodle.MoodleClient
}

type userResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	Firstname types.String `tfsdk:"firstname"`
	Lastname  types.String `tfsdk:"lastname"`
	Email     types.String `tfsdk:"email"`
	Auth      types.String `tfsdk:"auth"`
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Moodle user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The internal ID of the Moodle user.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The username of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The password of the user.",
			},
			"firstname": schema.StringAttribute{
				Required:    true,
				Description: "The first name of the user.",
			},
			"lastname": schema.StringAttribute{
				Required:    true,
				Description: "The last name of the user.",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "The email address of the user.",
			},
			"auth": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The authentication method of the user (default: manual).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := "manual"
	if !plan.Auth.IsNull() && !plan.Auth.IsUnknown() {
		auth = plan.Auth.ValueString()
	}

	tflog.Info(ctx, "Creating Moodle user", map[string]interface{}{
		"username": plan.Username.ValueString(),
	})

	user, err := r.client.CreateUser(
		plan.Username.ValueString(),
		plan.Password.ValueString(),
		plan.Firstname.ValueString(),
		plan.Lastname.ValueString(),
		plan.Email.ValueString(),
		auth,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	plan.ID = types.Int64Value(user.ID)
	// Password is not returned by API, so we keep what was in plan
	if user.Auth != "" {
		plan.Auth = types.StringValue(user.Auth)
	} else {
		plan.Auth = types.StringValue(auth)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(state.ID.ValueInt64())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	state.Username = types.StringValue(user.Username)
	state.Firstname = types.StringValue(user.Firstname)
	state.Lastname = types.StringValue(user.Lastname)
	state.Email = types.StringValue(user.Email)
	state.Auth = types.StringValue(user.Auth)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := "manual"
	if !plan.Auth.IsNull() && !plan.Auth.IsUnknown() {
		auth = plan.Auth.ValueString()
	}

	tflog.Info(ctx, "Updating Moodle user", map[string]interface{}{
		"user_id": state.ID.ValueInt64(),
	})

	err := r.client.UpdateUser(
		state.ID.ValueInt64(),
		plan.Username.ValueString(),
		plan.Password.ValueString(),
		plan.Firstname.ValueString(),
		plan.Lastname.ValueString(),
		plan.Email.ValueString(),
		auth,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}

	// Anchor state to what we sent so partial success is preserved in state
	// even if the subsequent verification step fails.
	plan.ID = state.ID
	plan.Auth = types.StringValue(auth)

	// Read back from Moodle to detect silent update failures (core_user_update_users
	// returns null even when Moodle ignores a field change, e.g. due to email
	// confirmation settings or missing capabilities).
	updated, err := r.client.GetUser(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError("Error verifying user update", err.Error())
		return
	}
	if updated.Email != plan.Email.ValueString() {
		// Store the actual email Moodle has so Terraform detects the ongoing drift
		// and plans another update on the next apply.
		wantedEmail := plan.Email.ValueString()
		plan.Email = types.StringValue(updated.Email)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError(
			"User email was not updated",
			fmt.Sprintf("Moodle did not apply the email change: got %q, want %q. "+
				"Check that the webservice token has 'moodle/user:update' capability and that "+
				"$CFG->emailchangeconfirmation is disabled.", updated.Email, wantedEmail),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(state.ID.ValueInt64())
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "accessexception") {
			tflog.Warn(ctx, "Moodle denied delete (accessexception); treating as already removed", map[string]interface{}{
				"user_id": state.ID.ValueInt64(),
			})
			return
		}
		if strings.Contains(errMsg, "not found") {
			tflog.Warn(ctx, "User already deleted out-of-band; removing from state", map[string]interface{}{
				"user_id": state.ID.ValueInt64(),
			})
			return
		}
		resp.Diagnostics.AddError("Error deleting user", errMsg)
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected a numeric Moodle user ID, got %q: %s", req.ID, err),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
