package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupChoiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupChoiceConfig("Pick Your Team"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_group_choice.test", "name", "Pick Your Team"),
					resource.TestCheckResourceAttr("moodle_group_choice.test", "section", "0"),
					resource.TestCheckResourceAttr("moodle_group_choice.test", "visible", "true"),
					resource.TestCheckResourceAttr("moodle_group_choice.test", "multipleenrollmentspossible", "false"),
					resource.TestCheckResourceAttr("moodle_group_choice.test", "showresults", "0"),
					resource.TestCheckResourceAttr("moodle_group_choice.test", "allowupdate", "false"),
					resource.TestCheckResourceAttrSet("moodle_group_choice.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_group_choice.test", "course_id"),
				),
			},
			{
				Config: testAccGroupChoiceConfig("Pick Your Team Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_group_choice.test", "name", "Pick Your Team Updated"),
				),
			},
		},
	})
}

func testAccGroupChoiceConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "moodle_course" "cg_course" {
  fullname   = "Group Choice Test Course"
  shortname  = "CG-TEST-01"
  categoryid = 1
}

resource "moodle_group" "cg_group_a" {
  course_id     = moodle_course.cg_course.id
  name          = "Group A"
  visibility    = 0
  participation = true
  enrolmentkey  = ""
  idnumber      = ""
}

resource "moodle_group" "cg_group_b" {
  course_id     = moodle_course.cg_course.id
  name          = "Group B"
  visibility    = 0
  participation = true
  enrolmentkey  = ""
  idnumber      = ""
}

resource "moodle_group_choice" "test" {
  course_id                   = moodle_course.cg_course.id
  section                     = 0
  name                        = %[1]q
  description                 = "<p>Choose your group.</p>"
  group_ids                   = [moodle_group.cg_group_a.id, moodle_group.cg_group_b.id]
  multipleenrollmentspossible = false
  showresults                 = 0
  allowupdate                 = false
  visible                     = true
  previous_element_id         = 0
}
`, name)
}
