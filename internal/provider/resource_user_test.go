package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_user.test", "username", "tf.test.user"),
					resource.TestCheckResourceAttr("moodle_user.test", "firstname", "Test"),
					resource.TestCheckResourceAttr("moodle_user.test", "lastname", "User"),
					resource.TestCheckResourceAttr("moodle_user.test", "email", "tf.test.user@example.com"),
					resource.TestCheckResourceAttr("moodle_user.test", "auth", "manual"),
					resource.TestCheckResourceAttrSet("moodle_user.test", "id"),
				),
			},
		},
	})
}

func testAccUserConfig() string {
	return providerConfig + `
resource "moodle_user" "test" {
  username  = "tf.test.user"
  password  = "TestPass123!"
  firstname = "Test"
  lastname  = "User"
  email     = "tf.test.user@example.com"
  auth      = "manual"
}
`
}
