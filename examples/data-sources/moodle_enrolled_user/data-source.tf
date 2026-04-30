data "moodle_enrolled_user" "course_participants" {
  course_id = moodle_course.example.id
}
