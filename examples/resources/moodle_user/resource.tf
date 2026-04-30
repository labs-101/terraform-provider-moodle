resource "moodle_user" "student_jane" {
  username  = "jane.doe"
  password  = "SecurePass123!"
  firstname = "Jane"
  lastname  = "Doe"
  email     = "jane.doe@example.com"
  auth      = "manual"
}
