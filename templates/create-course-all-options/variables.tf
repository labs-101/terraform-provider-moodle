variable "moodle_token" {
  description = "Moodle API token"
  type        = string
}

variable "student_emails" {
  description = "List of emails of the students to enrol"
  type        = list(string)
  default     = ["danielfelixstoecklein@gmail.com"]
}

variable "student_role_id" {
  description = "Role ID for the student (e.g., 5 for student)"
  type        = number
  default     = 5
}

variable "gitea_url" {
  description = "URL of the Gitea instance"
  type        = string
  default     = "https://gitea.project101.tech"
}

variable "gitea_username" {
  description = "Username for Gitea admin"
  type        = string
  default     = "gitea_admin"
}

variable "gitea_password" {
  description = "Password for Gitea admin"
  type        = string
  sensitive   = true
}