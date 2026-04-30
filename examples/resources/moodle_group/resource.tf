resource "moodle_group" "team_one" {
  course_id     = moodle_course.example.id
  name          = "Team One"
  description   = "<p>First project group.</p>"
  enrolmentkey  = ""
  visibility    = 0
  participation = 1
  idnumber      = ""
}
