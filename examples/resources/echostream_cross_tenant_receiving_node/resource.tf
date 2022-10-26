resource "echostream_cross_tenant_receiving_app" "test" {
  name           = "from_other_tenant"
  sending_tenant = "other_tenant"
}

resource "echostream_cross_tenant_receiving_node" "test" {
  name = "other_tenant:other_tenant_sending_node"
}
