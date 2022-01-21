# terraform-provider-git &nbsp; [![Build Status](https://github.com/au2001/terraform-provider-git/workflows/release/badge.svg)](https://github.com/au2001/terraform-provider-git/actions)

A [Terraform](http://terraform.io) plugin to manage files in Git repositories.

Available on the Terraform registry as [au2001/git](https://registry.terraform.io/providers/au2001/git).

## Installation

```hcl
terraform {
  required_providers {
    git = {
      source  = "au2001/git"
      version = "~> 0.1"
    }
  }
}
```

## Usage

```hcl
# Define your Git repository and credentials in case of a private repository
data "git_repository" "example" {
  url = "https://example.com/repo-name"
  ref = "main"
}

# Create a file in the Git repository without pushing
resource "git_file" "hello_world" {
  path    = "path/to/file.txt"
  content = "Hello, World!"
}

# Commit your changes to the Git repository
resource "git_commit" "hello_world" {
  message = "Create file.txt"
}

# Read a file in the Git repository
data "git_file" "read_hello_world" {
  path = "path/to/file.txt"
}

output "hello_world" {
  value = data.git_file.read_hello_world.content
}
```

## Resources

### git_file

### git_commit

## Data Sources

### git_repository

### git_file

# License

Licensed under the Apache License, Version 2.0.\
See the included LICENSE file for more details.
