package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupMemberResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMemberConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("moodle_group_member.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_group_member.test", "group_id"),
					resource.TestCheckResourceAttrSet("moodle_group_member.test", "user_id"),
				),
			},
		},
	})
}

func testAccGroupMemberConfig() string {
	return providerConfig + `
resource "moodle_course" "gm_course" {
  fullname   = "Group Member Test Course"
  shortname  = "GM-TEST-01"
  categoryid = 1
}

resource "moodle_user" "gm_user" {
  username  = "tf.gm.user"
  password  = "TestPass123!"
  firstname = "Group"
  lastname  = "Member"
  email     = "tf.gm.user@example.com"
  auth      = "manual"
}

resource "moodle_user_enrolment" "gm_enrol" {
  user_email = moodle_user.gm_user.email
  course_id  = moodle_course.gm_course.id
  role_id    = 5
}

resource "moodle_group" "gm_group" {
  course_id     = moodle_course.gm_course.id
  name          = "Test Group"
  visibility    = 0
  participation = 1
  enrolmentkey  = ""
  idnumber      = ""
}

resource "moodle_group_member" "test" {
  group_id   = moodle_group.gm_group.id
  user_id    = moodle_user.gm_user.id
  depends_on = [moodle_user_enrolment.gm_enrol]
}
`
}
