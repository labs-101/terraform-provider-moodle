resource "moodle_quiz" "midterm" {
  course_id         = moodle_course.example.id
  name              = "Midterm Exam"
  intro             = "<p>This quiz covers chapters 1-5.</p>"
  password          = "secret123"
  timeopen          = "2026-06-15"
  timeclose         = "2026-06-16"
  timelimit         = 3600
  attempts          = 1
  grademethod       = 1
  questionsperpage  = 5
  navmethod         = "sequential"
  section           = 1
  visible           = 1
}
