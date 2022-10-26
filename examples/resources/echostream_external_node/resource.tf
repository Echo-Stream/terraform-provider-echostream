data "echostream_message_type" "json" {
  name = "echo.json"
}

resource "echostream_external_app" "test" {
  name = "test"
}

resource "echostream_external_node" "test" {
  app                  = echostream_external_app.test.name
  name                 = "test"
  receive_message_type = data.echostream_message_type.json.name
  send_message_type    = data.echostream_message_type.json.name
}
