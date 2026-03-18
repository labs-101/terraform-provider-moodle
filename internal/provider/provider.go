// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"terraform-moodle-provider/internal/moodle"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &MoodleProvider{}
var _ provider.ProviderWithFunctions = &MoodleProvider{}
var _ provider.ProviderWithEphemeralResources = &MoodleProvider{}
var _ provider.ProviderWithActions = &MoodleProvider{}

type MoodleProvider struct {
	version string
}

type MoodleProviderModel struct {
	Host          types.String `tfsdk:"host"`
	Token         types.String `tfsdk:"token"`
	MoodleVersion types.String `tfsdk:"moodle_version"`
}

func (p *MoodleProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "moodle"
	resp.Version = p.version
}

func (p *MoodleProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required: true,
			},
			"token": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"moodle_version": schema.StringAttribute{
				Optional:    true,
				Description: "Die Moodle-Version der Zielinstanz (z.B. \"4.3\").",
			},
		},
	}
}

func (p *MoodleProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data MoodleProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Moodle API Host",
			"The provider cannot create the Moodle API client as there is an unknown configuration value for the Moodle API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_HOST environment variable.",
		)
	}

	if data.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Moodle API token",
			"The provider cannot create the Moodle API client as there is an unknown configuration value for the Moodle API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("MOODLE_HOST")
	token := os.Getenv("MOODLE_TOKEN")
	moodleVersion := os.Getenv("MOODLE_VERSION")

	if !data.Host.IsNull() {
		host = data.Host.ValueString()
	}

	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	if !data.MoodleVersion.IsNull() {
		moodleVersion = data.MoodleVersion.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Moodle API Host",
			"The provider cannot create the Moodle API client as there is a missing or empty value for the Moodle API host. "+
				"Set the host value in the configuration or use the HASHICUPS_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Moodle API Username",
			"The provider cannot create the Moodle API client as there is a missing or empty value for the Moodle API username. "+
				"Set the username value in the configuration or use the HASHICUPS_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := moodle.NewMoodleClient(host, token, moodleVersion)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Moodle API Client",
			"An unexpected error occurred when creating the Moodle API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"HashiCups Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *MoodleProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCourseResource,
		NewCourseSectionResource,
		NewSectionFileResource,
		NewSectionChoiceResource,
		NewSectionAssignmentResource,
		NewUserResource,
		NewUserEnrolmentResource,
	}
}

func (p *MoodleProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return nil
}

func (p *MoodleProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		GetAllCoursesDataSource,
	}
}

func (p *MoodleProvider) Functions(ctx context.Context) []func() function.Function {
	return nil
}

func (p *MoodleProvider) Actions(ctx context.Context) []func() action.Action {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MoodleProvider{
			version: version,
		}
	}
}
