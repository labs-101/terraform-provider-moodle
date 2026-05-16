resource "moodle_section_choice" "favourite_language" {
  course_id      = moodle_course.example.id
  section        = moodle_course_section.choice.section
  name           = "What is your favourite programming language?"
  description    = "<p>Please select one option.</p>"
  options        = ["Go", "Python", "Java", "TypeScript"]
  allow_multiple = false
}
