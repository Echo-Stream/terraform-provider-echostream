data "echostream_processor_function" "json_2_xml" {
  name = "echo.json:echo.xml"
}

data "echostream_message_type" "json" {
  name = "echo.json"
}

resource "echostream_processor_node" "test" {
  description           = "my processor node"
  managed_processor     = data.echostream_processor_function.json_2_xml.name
  name                  = "test"
  receive_message_type  = data.echostream_message_type.json.name
  send_message_type     = "echo.xml"
  sequential_processing = false
}
