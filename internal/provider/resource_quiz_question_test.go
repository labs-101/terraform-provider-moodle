package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccQuizQuestionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQuizQuestionConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("moodle_quiz_question.test", "name", "Capital of Germany"),
					resource.TestCheckResourceAttr("moodle_quiz_question.test", "type", "multichoice"),
					resource.TestCheckResourceAttr("moodle_quiz_question.test", "question_text", "What is the capital of Germany?"),
					resource.TestCheckResourceAttrSet("moodle_quiz_question.test", "id"),
					resource.TestCheckResourceAttrSet("moodle_quiz_question.test", "slot"),
					resource.TestCheckResourceAttrSet("moodle_quiz_question.test", "page"),
				),
			},
		},
	})
}

func testAccQuizQuestionConfig() string {
	return providerConfig + `
resource "moodle_course" "qq_course" {
  fullname   = "Quiz Question Test Course"
  shortname  = "QQ-TEST-01"
  categoryid = 1
}

resource "moodle_quiz" "qq_quiz" {
  course_id = moodle_course.qq_course.id
  name      = "Test Quiz"
}

resource "moodle_quiz_question" "test" {
  course_id     = moodle_course.qq_course.id
  quiz_id       = moodle_quiz.qq_quiz.id
  name          = "Capital of Germany"
  type          = "multichoice"
  question_text = "What is the capital of Germany?"

  choice {
    answer   = "Berlin"
    grade    = 1.0
    feedback = "Correct!"
  }

  choice {
    answer   = "Munich"
    grade    = 0.0
    feedback = "Munich is not the capital."
  }

  choice {
    answer   = "Hamburg"
    grade    = 0.0
    feedback = "Hamburg is not the capital."
  }
}
`
}
