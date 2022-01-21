terraform {
  required_providers {
    git = {
      source  = "au2001/git"
      version = "0.0.0-dev"
    }
  }
}

data "git_repository" "basic" {
  id  = "fixtures"
  url = "https://github.com/git-fixtures/basic.git"
  ref = "master"
}

data "git_file" "short_json" {
  repository = data.git_repository.basic.id
  path       = "./json/short.json"
}

output "short_json" {
  value = jsondecode(data.git_file.short_json.content)
}
