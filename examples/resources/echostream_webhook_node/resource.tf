resource "echostream_webhook_node" "test" {
  name              = "test"
  send_message_type = "echo.fhir-json"
}
