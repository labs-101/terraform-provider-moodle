package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSectionFileResource(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "upload.txt")
	if err := os.WriteFile(tmpFile, []byte("Terraform provider acceptance test file."), 0644); err != nil {
		t.Fatalf("could not create temp test file: %s", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSectionFileConfig(tmpFile),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_section_file.test", "display_name", "Test Document"),
					resource.TestCheckResourceAttr("moodle_section_file.test", "section", "0"),
					resource.TestCheckResourceAttr("moodle_section_file.test", "visible", "true"),
					resource.TestCheckResourceAttrSet("moodle_section_file.test", "id"),
				),
			},
		},
	})
}

func testAccSectionFileConfig(filePath string) string {
	return providerConfig + fmt.Sprintf(`
resource "moodle_course" "file_course" {
  fullname   = "File Test Course"
  shortname  = "FILE-TEST-01"
  categoryid = 1
}

resource "moodle_section_file" "test" {
  course_id    = moodle_course.file_course.id
  section  = 0
  file_path    = %[1]q
  display_name = "Test Document"
  visible      = true
}
`, filePath)
}
