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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &userResource{}
	_ resource.ResourceWithConfigure = &userResource{}
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
	if !plan.Auth.IsNull() {
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
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	state.Username = types.StringValue(user.Username)
	state.Firstname = types.StringValue(user.Firstname)
	state.Lastname = types.StringValue(user.Lastname)
	state.Email = types.StringValue(user.Email)
	state.Auth = types.StringValue(user.Auth)
	// Password cannot be read back

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// core_user_update_users is not implemented in moodle client yet, and moodle create user usually doesn't update.
	// For now we can return error or leave it empty if we don't support update.
	// Or we can assume forces replacement if some fields change, but schema handles that if we set ForceNew on attributes.
	// We didn't set ForceNew, so Terraform assumes Update is possible.
	// Since I didn't implement UpdateUser in client, I should probably fail or assume it's not supported.
	// But let's check if I can just implement it quickly or skip it.
	// Given the prompt "add 10 students", Create is the most important.
	// I'll leave Update empty but it might confuse Terraform if state changes.
	// Actually, I should probably implement UpdateUser in user.go if I want to be compliant,
	// but for this task I will just warn that update is not supported or implement it.
	// Let's implement UpdateUser in user.go later if needed. For now, I'll return an error saying not supported.
	resp.Diagnostics.AddError("Error updating user", "Update operation is not yet supported for Moodle users.")
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
		return
	}
}
