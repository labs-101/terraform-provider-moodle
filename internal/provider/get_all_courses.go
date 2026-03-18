package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-moodle-provider/internal/moodle"
)

var _ datasource.DataSource = &coursesDataSource{}
var _ datasource.DataSourceWithConfigure = &coursesDataSource{}

func GetAllCoursesDataSource() datasource.DataSource {
	return &coursesDataSource{}
}

type coursesDataSourceModel struct {
	Id      types.String  `tfsdk:"id"`
	Courses []courseModel `tfsdk:"courses"`
}

type coursesDataSource struct {
	client *moodle.MoodleClient
}

type courseModel struct {
	Id        types.Int64  `tfsdk:"id"`
	Shortname types.String `tfsdk:"shortname"`
	Fullname  types.String `tfsdk:"fullname"`
}

func (d *coursesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_courses"
}

func (d *coursesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"courses": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"shortname": schema.StringAttribute{
							Computed: true,
						},
						"fullname": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *coursesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*moodle.MoodleClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unerwarteter Provider-Typ",
			"Erwartete *MoodleClient. Bitte melde diesen Fehler.",
		)
		return
	}

	d.client = client
}

func (d *coursesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state coursesDataSourceModel

	courses, err := d.client.GetAllCourses()
	if err != nil {
		resp.Diagnostics.AddError(
			"Fehler beim Abrufen der Moodle-Kurse",
			"Die Moodle API hat einen Fehler zurückgegeben: "+err.Error(),
		)
		return
	}

	state.Id = types.StringValue("all-moodle-courses")

	for _, course := range courses {
		state.Courses = append(state.Courses, courseModel{
			Id:        types.Int64Value(course.Id),
			Shortname: types.StringValue(course.Shortname),
			Fullname:  types.StringValue(course.Fullname),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
