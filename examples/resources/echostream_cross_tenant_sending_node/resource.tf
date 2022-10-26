data "echostream_processor_function" "json_2_xml" {
  name = "echo.json:echo.xml"
}

data "echostream_message_type" "json" {
  name = "echo.json"
}

resource "echostream_cross_tenant_sending_app" "test" {
  name             = "sending_app"
  receiving_app    = "receiving_app"
  receiving_tenant = "receiving_tenant"
}

resource "echostream_cross_tenant_sending_node" "test" {
  app                   = echostream_cross_tenant_sending_app.test.name
  description           = "my sending node"
  managed_processor     = data.echostream_processor_function.json_2_xml.name
  name                  = "test"
  receive_message_type  = data.echostream_message_type.json.name
  send_message_type     = "echo.xml"
  sequential_processing = true
}
