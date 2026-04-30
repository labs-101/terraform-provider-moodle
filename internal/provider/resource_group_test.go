package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig("Team Alpha"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_group.test", "name", "Team Alpha"),
					resource.TestCheckResourceAttr("moodle_group.test", "description", "<p>First group.</p>"),
					resource.TestCheckResourceAttr("moodle_group.test", "visibility", "0"),
					resource.TestCheckResourceAttr("moodle_group.test", "participation", "1"),
					resource.TestCheckResourceAttrSet("moodle_group.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_group.test", "course_id"),
				),
			},
			{
				Config: testAccGroupConfig("Team Alpha Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_group.test", "name", "Team Alpha Updated"),
				),
			},
		},
	})
}

func testAccGroupConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "moodle_course" "grp_course" {
  fullname   = "Group Test Course"
  shortname  = "GRP-TEST-01"
  categoryid = 1
}

resource "moodle_group" "test" {
  course_id     = moodle_course.grp_course.id
  name          = %[1]q
  description   = "<p>First group.</p>"
  visibility    = 0
  participation = 1
  enrolmentkey  = ""
  idnumber      = ""
}
`, name)
}
