resource "moodle_section_label" "intro" {
  course_id   = moodle_course.example.id
  section     = moodle_course_section.label.section
  name        = "Welcome"
  description = "<h3>Welcome to the course!</h3><p>Please read the materials below before starting the assignment.</p>"
  visible     = 1
}
