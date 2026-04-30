resource "moodle_user_enrolment" "jane_in_course" {
  user_email = moodle_user.student_jane.email
  course_id  = moodle_course.example.id
  role_id    = 5 # 5 = student
}
