locals {
  tenant_config = {
    value_1 = "my value"
  }
}

resource "echostream_tenant" "current" {
  description = "This is my description"
  config      = jsonencode(locals.tenant_config)
}
