resource "moodle_course_section" "introduction" {
  course_id = moodle_course.example.id
  name      = "Introduction"
}
