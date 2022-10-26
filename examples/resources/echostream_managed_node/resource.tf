data "echostream_managed_node_type" "hl7_in" {
  name = "echo.hl7-mllp-inbound-node"
}

resource "echostream_managed_app" "test" {
  name = "test"
}

resource "echostream_managed_node" "test" {
  app               = echostream_managed_app.test.name
  managed_node_type = data.echostream_managed_node_type.hl7_in.name
  name              = "test"
  ports = [
    {
      container_port = 2575
      host_port      = 2575
      protocol       = "tcp"
    }
  ]
}
