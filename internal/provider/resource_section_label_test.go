package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSectionLabelResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSectionLabelConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_section_label.test", "description", "<p>Gültig ab dem <strong>18.04.2026</strong>.</p>"),
					resource.TestCheckResourceAttr("moodle_section_label.test", "section", "1"),
					resource.TestCheckResourceAttr("moodle_section_label.test", "visible", "true"),
					resource.TestCheckResourceAttrSet("moodle_section_label.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_section_label.test", "course_id"),
				),
			},
			{
				Config: testAccSectionLabelConfigUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_section_label.test", "description", "<p>Aktualisiert: Gültig ab dem <strong>18.04.2026</strong>.</p>"),
					resource.TestCheckResourceAttr("moodle_section_label.test", "visible", "true"),
					resource.TestCheckResourceAttrSet("moodle_section_label.test", "id"),
				),
			},
		},
	})
}

func testAccSectionLabelConfig() string {
	return providerConfig + `
resource "moodle_course" "label_course" {
  fullname   = "Label Test Course"
  shortname  = "LABEL-TEST-01"
  categoryid = 1
}

resource "moodle_section_label" "test" {
  course_id   = moodle_course.label_course.id
  section     = 1
  description = "<p>Gültig ab dem <strong>18.04.2026</strong>.</p>"
  visible     = true
}
`
}

func testAccSectionLabelConfigUpdated() string {
	return providerConfig + `
resource "moodle_course" "label_course" {
  fullname   = "Label Test Course"
  shortname  = "LABEL-TEST-01"
  categoryid = 1
}

resource "moodle_section_label" "test" {
  course_id   = moodle_course.label_course.id
  section     = 1
  description = "<p>Aktualisiert: Gültig ab dem <strong>18.04.2026</strong>.</p>"
  visible     = true
}
`
}
