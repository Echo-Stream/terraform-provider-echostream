resource "echostream_external_app" "test" {
  name = "test"
}

data "echostream_app_change_receiver_node" "test_receiver" {
  app = echostream_external_app.test.name
}
