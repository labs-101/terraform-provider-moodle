# Example of managing a moodle course
resource "moodle_course" "test_kurs" {
  fullname   = "Example course"
  shortname  = "ec"
  startdate  = "2026-03-10"
  categoryid = 1
  idnumber   = 2
  visibility = 1
  summary    = "This is an example course"
}
