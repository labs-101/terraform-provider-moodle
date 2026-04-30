resource "moodle_group_member" "jane_in_one" {
  group_id = moodle_group.team_one.id
  user_id  = moodle_user.student_jane.id
}
