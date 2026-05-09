resource "moodle_quiz" "midterm" {
  course_id = moodle_course.example.id
  section   = moodle_course_section.quiz.section
  name      = "Midterm Exam"
  intro     = "<p>This quiz covers chapters 1-5.</p>"
  timeopen  = "2026-06-15"
  timeclose = "2026-06-16"
  timelimit = 3600
  attempts  = 1
}
