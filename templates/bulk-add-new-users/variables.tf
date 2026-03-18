variable "moodle_token" {
  description = "Moodle API token"
  type        = string
  sensitive   = true
}

variable "moodle_host" {
  description = "Moodle Host URL"
  type        = string
  default     = "https://moodle.project101.tech"
}

variable "course_id" {
  description = "ID of the course to enrol students in"
  type        = number
}

variable "student_role_id" {
  description = "Role ID for the student (default: 5)"
  type        = number
  default     = 5
}
