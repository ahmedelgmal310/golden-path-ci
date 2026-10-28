output "bucket_arn" {
  description = "ARN of the bucket."
  value       = aws_s3_bucket.this.arn
}

output "bucket_name" {
  description = "Name of the bucket."
  value       = aws_s3_bucket.this.id
}

output "kms_key_arn" {
  description = "ARN of the customer-managed KMS key."
  value       = aws_kms_key.this.arn
}
