package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSectionChoiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSectionChoiceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_section_choice.test", "name", "Preferred Meeting Time"),
					resource.TestCheckResourceAttr("moodle_section_choice.test", "section", "0"),
					resource.TestCheckResourceAttr("moodle_section_choice.test", "allow_multiple", "false"),
					resource.TestCheckResourceAttr("moodle_section_choice.test", "options.#", "3"),
					resource.TestCheckResourceAttrSet("moodle_section_choice.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_section_choice.test", "course_id"),
				),
			},
		},
	})
}

func testAccSectionChoiceConfig() string {
	return providerConfig + `
resource "moodle_course" "choice_course" {
  fullname   = "Choice Test Course"
  shortname  = "CHC-TEST-01"
  categoryid = 1
}

resource "moodle_section_choice" "test" {
  course_id      = moodle_course.choice_course.id
  section    = 0
  name           = "Preferred Meeting Time"
  intro          = "<p>When can you meet?</p>"
  options        = ["Monday 10:00", "Tuesday 14:00", "Wednesday 16:00"]
  allow_multiple = false
}
`
}
