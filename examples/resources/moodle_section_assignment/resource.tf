resource "moodle_section_assignment" "project_submission" {
  course_id = moodle_course.example.id
  section   = moodle_course_section.assignment.section
  name      = "Project Submission"
  intro     = "<p>Submit your final project here.</p>"
  duedate   = "2026-07-31"
}
