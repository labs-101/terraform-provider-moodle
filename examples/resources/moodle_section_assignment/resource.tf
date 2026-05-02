resource "moodle_section_assignment" "project_submission" {
  course_id                 = moodle_course.example.id
  section               = 2
  name                      = "Project Submission"
  intro                     = "<p>Submit your final project here.</p>"
  duedate                   = "2026-07-31"
  allowsubmissionsfromdate  = "2026-06-01"
  maxbytes                  = 10485760
  maxfilesubmissions        = 3
  submissiontypes           = "file"
}
