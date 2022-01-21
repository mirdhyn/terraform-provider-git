# terraform-provider-git

[![Build Status](https://github.com/au2001/terraform-provider-git/workflows/release/badge.svg)](https://github.com/au2001/terraform-provider-git/actions)

## Synopsis

A [Terraform](http://terraform.io) plugin to manage files in Git repositories.

## Example:

```hcl
resource "git_repository" "example" {
  url    = "ssh://git@yourcompany.com/example"
  branch = "main"
}

resource "git_file" "hello_world" {
  path = "path/to/file.txt"
  contents = "Hello, World!"
}

resource "git_commit" "hello_world" {
  message = "Create file.txt"
}
```

## Resources

### git_repository

### git_file

### git_commit

## Data Sources

### git_file

# License

Apache2 - See the included LICENSE file for more details.
