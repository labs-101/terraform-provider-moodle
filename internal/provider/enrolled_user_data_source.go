package provider

import (
	"context"
	"fmt"
	"terraform-moodle-provider/internal/moodle"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &enrolledUserDataSource{}
	_ datasource.DataSourceWithConfigure = &enrolledUserDataSource{}
)

// NewEnrolledUserDataSource is a helper function to simplify the provider implementation.
func NewEnrolledUserDataSource() datasource.DataSource {
	return &enrolledUserDataSource{}
}

// enrolledUserDataSource is the data source implementation.

type enrolledUserDataSource struct {
	client *moodle.MoodleClient
}

type userEnrolmentDatasourceModel struct {
	ID        types.String  `tfsdk:"id"`
	UserEmail types.String  `tfsdk:"user_email"`
	CourseID  types.Int64   `tfsdk:"course_id"`
	RoleIDs   []types.Int64 `tfsdk:"role_ids"`
}

type usersEnrolmentResourceModel struct {
	CourseID types.Int64                    `tfsdk:"course_id"`
	Users    []userEnrolmentDatasourceModel `tfsdk:"users"`
}

// Metadata returns the data source type name.
func (d *enrolledUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enrolled_user"
}

// Schema defines the schema for the data source.
func (d *enrolledUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a list of enrolled Moodle users for a specific course.",
		Attributes: map[string]schema.Attribute{
			"course_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course to fetch users for.",
			},
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the enrolment (composite key).",
						},
						"user_email": schema.StringAttribute{
							Computed:    true,
							Description: "The mail address of the user.",
						},
						"course_id": schema.Int64Attribute{
							Computed:    true,
							Description: "The ID of the course.",
						},
						"role_ids": schema.ListAttribute{
							Computed:    true,
							ElementType: types.Int64Type,
							Description: "The IDs of the roles.",
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *enrolledUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersEnrolmentResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	users, err := d.client.GetEnrolledUsers(state.CourseID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading enrolled users", err.Error())
		return
	}

	var userStates []userEnrolmentDatasourceModel
	for _, apiUser := range users {
		roleIds := make([]types.Int64, len(users))

		for i, role := range apiUser.Roles {
			roleIds[i] = types.Int64Value(role.RoleID)
		}
		userStates = append(userStates, userEnrolmentDatasourceModel{
			ID:        types.StringValue(fmt.Sprintf("%d", apiUser.ID)),
			UserEmail: types.StringValue(apiUser.Email),
			CourseID:  types.Int64Value(state.CourseID.ValueInt64()),
			RoleIDs:   roleIds,
		})
	}

	state.Users = userStates
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *enrolledUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*moodle.MoodleClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *moodle.MoodleClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}
