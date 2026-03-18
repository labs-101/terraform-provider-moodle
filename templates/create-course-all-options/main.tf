terraform {
  required_providers {
    moodle = {
      source = "lokal/dev/moodle"
    }
    gitea = {
      source  = "go-gitea/gitea"
      version = "0.6.0"
    }
  }
}

provider "moodle" {
  host           = "https://moodle.project101.tech"
  token          = var.moodle_token
  moodle_version = "4.0"
}

provider "gitea" {
  base_url = var.gitea_url
  username = var.gitea_username
  password = var.gitea_password
  insecure = true
}


resource "moodle_course" "test_kurs" {
  fullname   = "Vollständiger Moolde Kurs"
  shortname  = "VMK-101-6"
  startdate  = "2026-03-10"
  categoryid = 1
  idnumber   = 6
  visibility = 1 # 1 = visible
  summary    = "Das ist mein erster vollständiger Moolde Kurs"
}

resource "moodle_course_section" "woche_1" {
  course_id = moodle_course.test_kurs.id
  name      = "Woche 1: Einführung in Terraform1"
}

resource "moodle_course_section" "woche_2" {
  course_id = moodle_course.test_kurs.id
  name      = "Woche 2: Einführung in Terraform1"
}

resource "moodle_section_file" "expose_daniel_stoecklein" {
  course_id    = moodle_course.test_kurs.id
  section_num  = moodle_course_section.woche_1.section
  file_path    = "${path.module}/Expose-Daniel-Stoecklein.pdf"
  file_hash    = filemd5("${path.module}/Expose-Daniel-Stoecklein.pdf")
  display_name = "Einführung (PDF)"
  visible      = 1
}

resource "moodle_section_file" "expose_test_txt" {
  course_id    = moodle_course.test_kurs.id
  section_num  = moodle_course_section.woche_1.section
  file_path    = "${path.module}/test.txt"
  file_hash    = filemd5("${path.module}/test.txt")
  display_name = "Text Datei"
  visible      = 1
}

resource "moodle_section_choice" "lieblingssprache" {
  course_id      = moodle_course.test_kurs.id
  section_num    = moodle_course_section.woche_1.section
  name           = "Welche Programmiersprache bevorzugst du1?"
  intro          = "<p>Bitte wähle deine bevorzugte Sprache aus.</p>"
  options        = ["Python", "Go", "Java", "PHP"]
  allow_multiple = false
}

resource "moodle_section_assignment" "aufgabe_1" {
  course_id                = moodle_course.test_kurs.id
  section_num              = moodle_course_section.woche_2.section
  name                     = "Aufgabe 1: Terraform Grundlagen"
  intro                    = "<p>Erstelle eine einfache Terraform-Konfiguration für einen Moodle-Kurs und dokumentiere deine Lösung.</p>"
  duedate                  = "2026-06-30"
  allowsubmissionsfromdate = "2026-03-11"
  submissiontypes          = "file,onlinetext"
  maxfilesubmissions       = 3
  maxbytes                 = 10485760
}

resource "moodle_user_enrolment" "students" {
  for_each   = toset(var.student_emails)
  user_email = each.value
  course_id  = moodle_course.test_kurs.id
  role_id    = var.student_role_id
}

resource "gitea_repository" "course_repo" {
  username = var.gitea_username
  name     = lower(replace(moodle_course.test_kurs.fullname, " ", "-"))
  private  = true
}

resource "gitea_repository_collaborator" "student_collabs" {
  for_each   = toset(var.student_emails)
  repository = gitea_repository.course_repo.name
  username   = split("@", each.value)[0]
  permission = "write"
}

