terraform {
  required_providers {
     moodle = {
      source = "lokal/dev/moodle"
    }
  }
}

provider "moodle" {
  host           = "https://moodle.project101.tech"
  token          = var.moodle_token
  moodle_version = "4.0"
}

data "moodle_enrolled_user" "users" {
    course_id = "9"
}

output "enrolled_users" {
  value = data.moodle_enrolled_user.users
}