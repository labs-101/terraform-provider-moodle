resource "moodle_section_choice" "favourite_language" {
  course_id      = moodle_course.example.id
  section    = 1
  name           = "What is your favourite programming language?"
  intro          = "<p>Please select one option.</p>"
  options        = ["Go", "Python", "Java", "TypeScript"]
  allow_multiple = false
}
