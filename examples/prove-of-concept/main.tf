terraform {
  required_providers {
    moodle = {
      source = "lokal/dev/moodle"
    }
  }
}

provider "moodle" {
   host  = "https://moodle.project101.tech"
   token = "a5f72cc9b7edf1567c14923a569d41c3"
}

data "moodle_courses" "example" {}

output "courses_test" {
  value = data.moodle_courses.example.courses
}