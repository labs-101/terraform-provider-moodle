package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccChoicegroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccChoicegroupConfig("Pick Your Team"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_choicegroup.test", "name", "Pick Your Team"),
					resource.TestCheckResourceAttr("moodle_choicegroup.test", "section_num", "0"),
					resource.TestCheckResourceAttr("moodle_choicegroup.test", "visible", "1"),
					resource.TestCheckResourceAttr("moodle_choicegroup.test", "multipleenrollmentspossible", "0"),
					resource.TestCheckResourceAttr("moodle_choicegroup.test", "showresults", "0"),
					resource.TestCheckResourceAttr("moodle_choicegroup.test", "allowupdate", "0"),
					resource.TestCheckResourceAttrSet("moodle_choicegroup.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_choicegroup.test", "course_id"),
				),
			},
			{
				Config: testAccChoicegroupConfig("Pick Your Team Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_choicegroup.test", "name", "Pick Your Team Updated"),
				),
			},
		},
	})
}

func testAccChoicegroupConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "moodle_course" "cg_course" {
  fullname   = "Choicegroup Test Course"
  shortname  = "CG-TEST-01"
  categoryid = 1
}

resource "moodle_group" "cg_group_a" {
  course_id     = moodle_course.cg_course.id
  name          = "Group A"
  visibility    = 0
  participation = 1
  enrolmentkey  = ""
  idnumber      = ""
}

resource "moodle_group" "cg_group_b" {
  course_id     = moodle_course.cg_course.id
  name          = "Group B"
  visibility    = 0
  participation = 1
  enrolmentkey  = ""
  idnumber      = ""
}

resource "moodle_choicegroup" "test" {
  course_id                   = moodle_course.cg_course.id
  section_num                 = 0
  name                        = %[1]q
  intro                       = "<p>Choose your group.</p>"
  group_ids                   = [moodle_group.cg_group_a.id, moodle_group.cg_group_b.id]
  multipleenrollmentspossible = 0
  showresults                 = 0
  allowupdate                 = 0
  visible                     = 1
  previous_element_id         = 0
}
`, name)
}
