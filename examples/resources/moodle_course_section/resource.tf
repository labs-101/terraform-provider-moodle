resource "moodle_course_section" "introduction" {
  course_id = moodle_course.example.id
  name      = "Introduction"
  summary   = "<p>Welcome to the course!</p>"
  section   = 1
  visible   = 1
}
