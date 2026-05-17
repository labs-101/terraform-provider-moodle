resource "moodle_group_choice" "pick_your_team" {
  course_id                   = moodle_course.example.id
  section                     = moodle_course_section.choicegroup.section
  name                        = "Pick Your Project Team"
  description                 = "<p>Please select the group you want to join.</p>"
  group_ids                   = [moodle_group.team_alpha.id, moodle_group.team_beta.id]
  multipleenrollmentspossible = false
  showresults                 = true
  allowupdate                 = true
  timeopen                    = "2026-05-01"
  timeclose                   = "2026-05-15"
  visible                     = true
}
