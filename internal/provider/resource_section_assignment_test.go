package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSectionAssignmentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSectionAssignmentConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_section_assignment.test", "name", "Project Submission"),
					resource.TestCheckResourceAttr("moodle_section_assignment.test", "section_num", "0"),
					resource.TestCheckResourceAttr("moodle_section_assignment.test", "duedate", "2026-12-31"),
					resource.TestCheckResourceAttr("moodle_section_assignment.test", "maxbytes", "10485760"),
					resource.TestCheckResourceAttr("moodle_section_assignment.test", "submissiontypes", "file"),
					resource.TestCheckResourceAttrSet("moodle_section_assignment.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_section_assignment.test", "course_id"),
				),
			},
		},
	})
}

func testAccSectionAssignmentConfig() string {
	return providerConfig + `
resource "moodle_course" "assn_course" {
  fullname   = "Assignment Test Course"
  shortname  = "ASSN-TEST-01"
  categoryid = 1
}

resource "moodle_section_assignment" "test" {
  course_id       = moodle_course.assn_course.id
  section_num     = 0
  name            = "Project Submission"
  intro           = "<p>Submit your project here.</p>"
  duedate         = "2026-12-31"
  maxbytes        = 10485760
  submissiontypes = "file"
}
`
}
