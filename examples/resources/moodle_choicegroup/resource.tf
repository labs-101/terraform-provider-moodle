resource "moodle_choicegroup" "pick_your_team" {
  course_id                   = moodle_course.example.id
  section                 = 1
  name                        = "Pick Your Project Team"
  intro                       = "<p>Please select the group you want to join.</p>"
  group_ids                   = [moodle_group.team_alpha.id, moodle_group.team_beta.id]
  multipleenrollmentspossible = 0
  showresults                 = 1
  allowupdate                 = 1
  timeopen                    = "2026-05-01"
  timeclose                   = "2026-05-15"
  visible                     = 1
  previous_element_id         = 0
}
