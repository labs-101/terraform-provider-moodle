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
	_ resource.Resource              = &courseResource{}
	_ resource.ResourceWithConfigure = &courseResource{}
)

func NewCourseResource() resource.Resource {
	return &courseResource{}
}

type courseResource struct {
	client *moodle.MoodleClient
}

type courseResourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Fullname   types.String `tfsdk:"fullname"`
	Shortname  types.String `tfsdk:"shortname"`
	CategoryID types.Int64  `tfsdk:"categoryid"`
	Idnumber   types.String `tfsdk:"idnumber"`
	Summary    types.String `tfsdk:"summary"`
	Visibility types.Int64  `tfsdk:"visibility"`
	StartDate  types.String `tfsdk:"startdate"`
	EndDate    types.String `tfsdk:"enddate"`
}

func (r *courseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_course"
}

func (r *courseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *courseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Moodle course.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The internal ID of the Moodle course.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"fullname": schema.StringAttribute{
				Required:    true,
				Description: "The full name of the course.",
			},
			"shortname": schema.StringAttribute{
				Required:    true,
				Description: "The short name (abbreviation) of the course.",
			},
			"categoryid": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the course category in which the course should be created.",
			},
			"idnumber": schema.StringAttribute{
				Optional:    true,
				Description: "The ID number of the course.",
			},
			"summary": schema.StringAttribute{
				Optional:    true,
				Description: "The summary/description of the course.",
			},
			"visibility": schema.Int64Attribute{
				Optional:    true,
				Description: "The visibility of the course (1 = visible, 0 = hidden).",
			},
			"startdate": schema.StringAttribute{
				Optional:    true,
				Description: "The start date of the course in format YYYY-MM-DD (e.g. 2026-03-07).",
			},
			"enddate": schema.StringAttribute{
				Optional:    true,
				Description: "The end date of the course in format YYYY-MM-DD (e.g. 2026-12-31).",
			},
		},
	}
}

func (r *courseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan courseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to get plan for course creation", map[string]interface{}{
			"error": resp.Diagnostics,
		})
		return
	}

	startdate, err := parseDateToUnix(plan.StartDate.ValueString())
	if err != nil {
		tflog.Error(ctx, "Invalid start date", map[string]interface{}{
			"startdate": plan.StartDate.ValueString(),
			"error":     err.Error(),
		})
		resp.Diagnostics.AddError("Invalid start date", err.Error())
		return
	}
	enddate, err := parseDateToUnix(plan.EndDate.ValueString())
	if err != nil {
		tflog.Error(ctx, "Invalid end date", map[string]any{
			"enddate": plan.EndDate.ValueString(),
			"error":   err.Error(),
		})
		resp.Diagnostics.AddError("Invalid end date", err.Error())
		return
	}

	tflog.Info(ctx, "Creating Moodle course", map[string]any{
		"fullname":   plan.Fullname.ValueString(),
		"shortname":  plan.Shortname.ValueString(),
		"categoryid": plan.CategoryID.ValueInt64(),
		"idnumber":   plan.Idnumber.ValueString(),
		"summary":    plan.Summary.ValueString(),
		"visibility": plan.Visibility.ValueInt64(),
		"startdate":  startdate,
		"enddate":    enddate,
	})

	course, err := r.client.CreateCourse(plan.Fullname.ValueString(), plan.Shortname.ValueString(), plan.CategoryID.ValueInt64(), plan.Idnumber.ValueString(), plan.Summary.ValueString(), plan.Visibility.ValueInt64(), startdate, enddate)
	if err != nil {
		tflog.Error(ctx, "Error creating course", map[string]any{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError("Error creating course", err.Error())
		return
	}

	tflog.Info(ctx, "Successfully created course", map[string]interface{}{
		"course_id": course.Id,
		"fullname":  course.Fullname,
	})

	plan.ID = types.Int64Value(course.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *courseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state courseResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	course, err := r.client.GetCourse(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading course", err.Error())
		return
	}

	state.Fullname = types.StringValue(course.Fullname)
	state.Shortname = types.StringValue(course.Shortname)
	state.Idnumber = types.StringValue(course.Idnumber)
	state.Summary = types.StringValue(course.Summary)
	state.Visibility = types.Int64Value(course.Visibility)

	if startDateStr := unixToDate(course.StartDate); startDateStr == "" {
		state.StartDate = types.StringNull()
	} else {
		state.StartDate = types.StringValue(startDateStr)
	}

	if endDateStr := unixToDate(course.EndDate); endDateStr == "" {
		state.EndDate = types.StringNull()
	} else {
		state.EndDate = types.StringValue(endDateStr)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *courseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan courseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	startdate, err := parseDateToUnix(plan.StartDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid start date", err.Error())
		return
	}
	enddate, err := parseDateToUnix(plan.EndDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid end date", err.Error())
		return
	}

	err = r.client.UpdateCourse(
		plan.ID.ValueInt64(),
		plan.Fullname.ValueString(),
		plan.Shortname.ValueString(),
		plan.CategoryID.ValueInt64(),
		plan.Idnumber.ValueString(),
		plan.Summary.ValueString(),
		plan.Visibility.ValueInt64(),
		startdate,
		enddate,
	)

	if err != nil {
		resp.Diagnostics.AddError("Error updating course", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *courseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state courseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCourse(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting course", err.Error())
		return
	}
}
