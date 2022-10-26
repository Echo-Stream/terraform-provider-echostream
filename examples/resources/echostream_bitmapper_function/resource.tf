resource "echostream_bitmapper_function" "test" {
  argument_message_type = "echo.text"
  code                  = "def bitmapper(*, context, message, source, **kwargs):\n\n    import simplejson as json\n\n    bitmap = 0x0\n\n    return bitmap\n"
  description           = "my description"
  name                  = "test"
}