# terraform-provider-git &nbsp; [![Build Status](https://github.com/arl-sh/terraform-provider-git/workflows/release/badge.svg)](https://github.com/arl-sh/terraform-provider-git/actions)

A [Terraform](http://terraform.io) plugin to manage files in Git repositories.

Available on the Terraform registry as [arl-sh/git](https://registry.terraform.io/providers/arl-sh/git).

## Installation

```hcl
terraform {
  required_providers {
    git = {
      source  = "arl-sh/git"
      version = "~> 0.1"
    }
  }
}
```

Then run `terraform init`.

## Usage

```hcl
# Clone the Git repository
resource "git_repository" "example" {
  url = "https://example.com/repo-name"
  ref = "main"
}
```

```hcl
# Create or edit a file in the Git repository without pushing
resource "git_file" "hello_world" {
  repository = git_repository.example.dir
  path       = "path/to/file.txt"
  content    = "Hello, World!"
}
```

```hcl
# Commit and push your changes to the Git repository
resource "git_commit" "hello_world" {
  repository = git_repository.example.dir
  message    = "Create file.txt"

  add {
    path = git_file.hello_world.path
  }
}
```

```hcl
# Read an existing file in the Git repository
data "git_file" "read_hello_world" {
  repository = git_repository.example.dir
  path       = "path/to/file.txt"
}

output "hello_world" {
  value = data.git_file.read_hello_world.content
}
```

## Resources

### git_repository

### git_file

### git_commit

## Data Sources

### git_file

# License

Licensed under the Apache License, Version 2.0.\
See the included LICENSE file for more details.
