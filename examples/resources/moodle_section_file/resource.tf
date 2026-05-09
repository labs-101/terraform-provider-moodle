resource "moodle_section_file" "file_pdf" {
  course_id    = moodle_course.example.id
  section      = moodle_course_section.file.section
  file_path    = "/path/to/file.pdf"
  display_name = "File.pdf"
  visible      = 1
}
