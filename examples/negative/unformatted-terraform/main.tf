# Negative fixture: valid Terraform that is not `terraform fmt` clean
# (misaligned "=" and indentation). The terraform workflow must fail at fmt.
terraform {
  required_version = ">= 1.10"
}

resource "terraform_data" "example" {
    input = {
  name = "golden-path"
      owner =   "platform"
    }
}
