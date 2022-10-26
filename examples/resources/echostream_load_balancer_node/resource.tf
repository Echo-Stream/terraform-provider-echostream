resource "echostream_load_balancer_node" "test" {
  name                 = "test"
  receive_message_type = "echo.text"
}
