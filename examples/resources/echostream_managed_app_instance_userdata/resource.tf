resource "echostream_managed_app" "test" {
  name = "test_app"
}

resource "echostream_managed_app_instance_userdata" "test" {
  app  = echostream_managed_app.test.name
  name = "2022-10-25T14:51:31"
}
