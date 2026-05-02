resource "moodle_section_file" "syllabus" {
  course_id    = moodle_course.example.id
  section  = 0
  file_path    = "/path/to/syllabus.pdf"
  display_name = "Course Syllabus"
  visible      = 1
}
