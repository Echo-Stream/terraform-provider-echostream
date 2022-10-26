resource "echostream_cross_tenant_sending_app" "test" {
  name             = "sending_app"
  receiving_app    = "receiving_app"
  receiving_tenant = "receiving_tenant"
}
