variable "client_id" {}
variable "client_secret" {}

variable "tenant_id" {}

variable "subscription_id" {}

variable "kubernetes_version" {}

variable "agent_count" {
    default = 1
}

variable "ssh_authorized_key" {
    default = "~/.ssh/id_rsa.pub"
}

variable "dns_prefix" {
    default = "akstest"
}

variable "cluster_name" {
    default = "akstest"
}

variable "resource_group_name" {
    default = "azure-akstest"
}

variable location {
    default = "East US"
}