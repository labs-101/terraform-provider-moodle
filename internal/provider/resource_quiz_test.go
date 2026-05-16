package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccQuizResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQuizConfig("Midterm Exam"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_quiz.test", "name", "Midterm Exam"),
					resource.TestCheckResourceAttr("moodle_quiz.test", "timelimit", "3600"),
					resource.TestCheckResourceAttr("moodle_quiz.test", "attempts", "1"),
					resource.TestCheckResourceAttrSet("moodle_quiz.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_quiz.test", "course_id"),
					resource.TestCheckResourceAttrSet("moodle_quiz.test", "section"),
					resource.TestCheckResourceAttrSet("moodle_quiz.test", "visible"),
				),
			},
			{
				Config: testAccQuizConfig("Midterm Exam Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_quiz.test", "name", "Midterm Exam Updated"),
				),
			},
		},
	})
}

func testAccQuizConfig(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "moodle_course" "quiz_course" {
  fullname   = "Quiz Test Course"
  shortname  = "QUIZ-TEST-01"
  categoryid = 1
}

resource "moodle_quiz" "test" {
  course_id = moodle_course.quiz_course.id
  name      = %[1]q
  description     = "<p>Quiz introduction.</p>"
  timelimit = 3600
  attempts  = 1
}
`, name)
}
