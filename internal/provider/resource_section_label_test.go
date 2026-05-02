package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSectionLabelResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and verify
			{
				Config: testAccSectionLabelConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_section_label.test", "intro", "<p>Gültig ab dem <strong>18.04.2026</strong>.</p>"),
					resource.TestCheckResourceAttr("moodle_section_label.test", "section_num", "1"),
					resource.TestCheckResourceAttr("moodle_section_label.test", "visible", "1"),
					resource.TestCheckResourceAttrSet("moodle_section_label.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_section_label.test", "course_id"),
				),
			},
			// Update intro in-place (no replacement)
			{
				Config: testAccSectionLabelConfigUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_section_label.test", "intro", "<p>Aktualisiert: Gültig ab dem <strong>18.04.2026</strong>.</p>"),
					resource.TestCheckResourceAttr("moodle_section_label.test", "visible", "1"),
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
  section_num = 1
  intro       = "<p>Gültig ab dem <strong>18.04.2026</strong>.</p>"
  visible     = 1
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
  section_num = 1
  intro       = "<p>Aktualisiert: Gültig ab dem <strong>18.04.2026</strong>.</p>"
  visible     = 1
}
`
}
