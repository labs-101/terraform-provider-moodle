package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCourseSectionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCourseSectionConfig("Week 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_course_section.test", "name", "Week 1"),
					resource.TestCheckResourceAttr("moodle_course_section.test", "visible", "1"),
					resource.TestCheckResourceAttrSet("moodle_course_section.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_course_section.test", "course_id"),
					resource.TestCheckResourceAttrSet("moodle_course_section.test", "section"),
				),
			},
			{
				Config: testAccCourseSectionConfig("Week 1 Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_course_section.test", "name", "Week 1 Updated"),
				),
			},
		},
	})
}

func testAccCourseSectionConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "moodle_course" "sec_course" {
  fullname   = "Section Test Course"
  shortname  = "SEC-TEST-01"
  categoryid = 1
}

resource "moodle_course_section" "test" {
  course_id = moodle_course.sec_course.id
  name      = %[1]q
  visible   = 1
}
`, name)
}
