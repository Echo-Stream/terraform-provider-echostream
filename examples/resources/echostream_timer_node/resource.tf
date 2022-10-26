resource "echostream_timer_node" "test" {
  description         = "Every 15 minutes"
  name                = "test"
  schedule_expression = "15 * ? * * *"
}
