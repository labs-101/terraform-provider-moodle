package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCourseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Ressource erstellen und Attribute prüfen
			{
				Config: testAccCourseResourceConfig("test-course"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_course.test", "fullname", "test-course"),
					resource.TestCheckResourceAttr("moodle_course.test", "shortname", "VMK-101-0"),
					resource.TestCheckResourceAttr("moodle_course.test", "startdate", "2026-03-10"),
					resource.TestCheckResourceAttr("moodle_course.test", "categoryid", "1"),
					resource.TestCheckResourceAttr("moodle_course.test", "idnumber", "10000"),
					resource.TestCheckResourceAttr("moodle_course.test", "visibility", "1"),
					resource.TestCheckResourceAttr("moodle_course.test", "summary", "test summary"),
					resource.TestCheckResourceAttrSet("moodle_course.test", "id"),
				),
			},
		},
	})
}

func testAccCourseResourceConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "moodle_course" "test" {
  fullname   = "%[1]s"
  shortname  = "VMK-101-0"
  startdate  = "2026-03-10"
  categoryid = 1
  idnumber   = 10000
  visibility = 1 # 1 = visible
  summary    = "test summary"
}
`, name)
}
