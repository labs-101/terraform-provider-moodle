resource "moodle_group" "team_one" {
  course_id   = moodle_course.example.id
  name        = "Team One"
  description = "<p>First project group.</p>"
}
