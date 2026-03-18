terraform {
  required_providers {
    gitea = {
      source  = "go-gitea/gitea"
      version = "0.6.0"
    }
  }
}

provider "gitea" {
  base_url = var.gitea_url
 
  username = var.username
  password = var.password

  insecure = false
}
resource "gitea_repository" "test" {
  username     = var.username
  name         = "repository-test"
  private      = true
  issue_labels = "Default"
  license      = "MIT"
  gitignores   = "Go"
}

resource "gitea_user" "students" {
  count                = 10
  username             = "student${count.index + 1}"
  login_name           = "student${count.index + 1}"
  password             = "Password123!"
  email                = "student${count.index + 1}@example.com"
  must_change_password = false
  active               = true
  admin                = false
}