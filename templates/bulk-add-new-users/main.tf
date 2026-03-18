terraform {
  required_providers {
    moodle = {
      source = "lokal/dev/moodle"
    }
  }
}

provider "moodle" {
  host  = var.moodle_host
  token = var.moodle_token
}

resource "moodle_user" "students" {
  count     = 10
  username  = "student${count.index + 1}"
  password  = "Password123!"
  firstname = "Student"
  lastname  = "${count.index + 1}"
  email     = "student${count.index + 1}@example.com"
  auth      = "manual"
}

resource "moodle_user_enrolment" "student_enrolment" {
  count     = 10
  user_id   = moodle_user.students[count.index].id
  course_id = var.course_id
  role_id   = var.student_role_id
}
