package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnrolledUserDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnrolledUserDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.moodle_enrolled_user.test", "course_id"),
					resource.TestCheckResourceAttrSet("data.moodle_enrolled_user.test", "users.#"),
					resource.TestCheckResourceAttr("data.moodle_enrolled_user.test", "users.0.user_email", "tf.eu.user@example.com"),
				),
			},
		},
	})
}

func testAccEnrolledUserDataSourceConfig() string {
	return providerConfig + `
resource "moodle_course" "eu_course" {
  fullname   = "Enrolled User Test Course"
  shortname  = "EU-TEST-01"
  categoryid = 1
}

resource "moodle_user" "eu_user" {
  username  = "tf.eu.user"
  password  = "TestPass123!"
  firstname = "Enrolled"
  lastname  = "User"
  email     = "tf.eu.user@example.com"
  auth      = "manual"
}

resource "moodle_user_enrolment" "eu_enrol" {
  user_email = moodle_user.eu_user.email
  course_id  = moodle_course.eu_course.id
  role_id    = 5
}

data "moodle_enrolled_user" "test" {
  course_id  = moodle_course.eu_course.id
  depends_on = [moodle_user_enrolment.eu_enrol]
}
`
}
