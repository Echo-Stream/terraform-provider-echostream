resource "echostream_processor_function" "test" {
  argument_message_type = "echo.json"
  code                  = "def processor(*, context, message, source, **kwargs):\n\n    from csv import DictReader\n    from io import StringIO\n\n    import simplejson as json\n\n    return json.dumps(\n        [row for row in DictReader(StringIO(message))], separators=(\",\", \":\")\n    )\n"
  description           = "Test function"
  name                  = "test"
  readme                = "# echo.csv:echo.json\n\nConverts a CSV message into a single JSON array object, with each element of the array representing a row in the CSV message.\n\nThe CSV message **must** have a header row as the first row.\n"
  requirements = [
    "simplejson"
  ]
  return_message_type = echostream_message_type.test.name
}
