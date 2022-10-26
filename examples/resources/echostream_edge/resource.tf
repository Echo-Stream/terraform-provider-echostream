resource "echostream_timer_node" "timer" {
  name                = "timer"
  schedule_expression = "15 * ? * * *"
}

resource "echostream_load_balancer_node" "lb1" {
  name                 = "lb1"
  receive_message_type = "echo.timer"
}

resource "echostream_edge" "timer_to_lb1" {
  source = data.echostream_timer_node.timer.name
  target = echostream_load_balancer_node.lb1.name
}
