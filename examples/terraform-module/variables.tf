variable "name" {
  description = "Bucket name (globally unique)."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", var.name))
    error_message = "name must be a valid S3 bucket name (3-63 lowercase letters, digits, dots, hyphens)."
  }
}

variable "access_log_bucket" {
  description = "Existing bucket that receives server access logs."
  type        = string
}

variable "noncurrent_version_days" {
  description = "Days to keep noncurrent object versions before expiry."
  type        = number
  default     = 90

  validation {
    condition     = var.noncurrent_version_days >= 1
    error_message = "noncurrent_version_days must be at least 1."
  }
}

variable "tags" {
  description = "Tags applied to every resource."
  type        = map(string)
  default     = {}
}
