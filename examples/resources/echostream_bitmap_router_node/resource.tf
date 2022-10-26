data "echostream_message_type" "text" {
  name = "echo.text"
}

resource "echostream_bitmap_router_node" "test" {
  description          = "my bitmap router"
  name                 = "test"
  inline_bitmapper     = data.echostream_message_type.text.bitmapper_template
  receive_message_type = data.echostream_message_type.text.name
  requirements         = data.echostream_message_type.text.requirements
  route_table = {
    "0x1" = ["node1", "node2"]
    "0X2" = ["node3"]
    "0x3" = ["node4"]
  }
}
