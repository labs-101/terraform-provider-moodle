resource "moodle_quiz_question" "capital_city" {
  course_id     = moodle_course.example.id
  quiz_id       = moodle_quiz.midterm.id
  name          = "Capital of France"
  type          = "multichoice"
  question_text = "What is the capital city of France?"

  choice {
    answer   = "Paris"
    grade    = 1.0
    feedback = "Correct!"
  }

  choice {
    answer   = "London"
    grade    = 0.0
    feedback = "London is the capital of the United Kingdom."
  }

  choice {
    answer   = "Berlin"
    grade    = 0.0
    feedback = "Berlin is the capital of Germany."
  }
}
