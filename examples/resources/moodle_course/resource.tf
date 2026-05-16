# Example of managing a moodle course
resource "moodle_course" "example" {
  fullname    = "Example Course"
  shortname   = "ec"
  startdate   = "2026-03-10"
  categoryid  = 1
  visibility  = 1
  description = "This is an example course."
}
