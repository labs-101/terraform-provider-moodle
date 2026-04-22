# Terraform Provider for Moodle

> **⚠ Work in Progress** — This provider is under active development. Breaking changes may occur at any time.

A Terraform provider for managing Moodle resources such as courses, sections, quizzes, users, and enrolments.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- A running Moodle instance with the **[Moodle Course API](../moodle-course-api/README.md)** plugin installed

> **Note:** This provider depends on the [Moodle Course API](../moodle-course-api/README.md) plugin. The plugin must be installed and enabled on your Moodle instance before using this provider, as it exposes the REST endpoints required by the provider.

## Provider Configuration

```hcl
provider "moodle" {
  host           = "https://moodle.example.com"
  token          = "moodle_service_token"
  moodle_version = "4.0"
}
```

| Argument          | Description                                      | Required |
|-------------------|--------------------------------------------------|----------|
| `host`            | Base URL of the Moodle instance                  | Yes      |
| `token`           | Moodle web service token                         | Yes      |
| `moodle_version`  | Moodle version (e.g. `"4.0"`)                    | Yes      |

The `token` and `host` values can also be set via the environment variables `MOODLE_TOKEN` and `MOODLE_HOST`.

## Supported Resources

| Resource                        | Description                              |
|---------------------------------|------------------------------------------|
| `moodle_course`                 | Manage Moodle courses                    |
| `moodle_course_section`         | Manage sections within a course          |
| `moodle_section_assignment`     | Add assignments to a course section      |
| `moodle_section_choice`         | Add choice activities to a section       |
| `moodle_section_file`           | Upload files to a section                |
| `moodle_quiz`                   | Manage quizzes                           |
| `moodle_quiz_question`          | Manage quiz questions                    |
| `moodle_user`                   | Manage Moodle users                      |
| `moodle_user_enrolment`         | Enrol users into courses                 |

## Supported Data Sources

| Data Source              | Description                              |
|--------------------------|------------------------------------------|
| `moodle_enrolled_users`  | Retrieve users enrolled in a course      |

## Building from Source

```bash
go build -o terraform-provider-moodle
```