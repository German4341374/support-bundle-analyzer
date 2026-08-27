variable "image" {
  description = "Prebuilt Support Bundle Analyzer image."
  type        = string
  default     = "ghcr.io/german4341374/support-bundle-analyzer:0.1.0"
  validation {
    condition     = length(var.image) > 3 && !endswith(var.image, ":latest")
    error_message = "Use an explicit immutable release tag instead of latest."
  }
}

variable "host_port" {
  description = "Loopback port for the local API."
  type        = number
  default     = 8080
  validation {
    condition     = var.host_port >= 1024 && var.host_port <= 65535
    error_message = "host_port must be between 1024 and 65535."
  }
}

variable "access_token" {
  description = "Bearer token used when the container listens outside its own loopback interface."
  type        = string
  sensitive   = true
  validation {
    condition     = length(var.access_token) >= 24
    error_message = "access_token must contain at least 24 characters."
  }
}
