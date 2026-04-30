package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserEnrolmentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserEnrolmentConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_user_enrolment.test", "user_email", "tf.enrl.user@example.com"),
					resource.TestCheckResourceAttr("moodle_user_enrolment.test", "role_id", "5"),
					resource.TestCheckResourceAttrSet("moodle_user_enrolment.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_user_enrolment.test", "course_id"),
				),
			},
		},
	})
}

func testAccUserEnrolmentConfig() string {
	return providerConfig + `
resource "moodle_course" "enrl_course" {
  fullname   = "Enrolment Test Course"
  shortname  = "ENR-TEST-01"
  categoryid = 1
}

resource "moodle_user" "enrl_user" {
  username  = "tf.enrl.user"
  password  = "TestPass123!"
  firstname = "Enrl"
  lastname  = "User"
  email     = "tf.enrl.user@example.com"
  auth      = "manual"
}

resource "moodle_user_enrolment" "test" {
  user_email = moodle_user.enrl_user.email
  course_id  = moodle_course.enrl_course.id
  role_id    = 5
}
`
}
