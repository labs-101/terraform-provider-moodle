terraform {
  required_providers {
    moodle = {
      source = "lokal/dev/moodle"
    }
  }
}

provider "moodle" {
   host           = "https://moodle.project101.tech"
   token          = "a5f72cc9b7edf1567c14923a569d41c3"
   moodle_version = "4.3"
}

resource "moodle_course" "test_kurs" {
  fullname   = "Mein erster Terraform Kurs2"
  shortname  = "TF-101"
  startdate  = "2026-03-10"
  enddate    = "2027-03-10"
  categoryid = 1
}
