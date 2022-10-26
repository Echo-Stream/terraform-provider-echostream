locals {
  mnt_config = {
    additionalProperties = false,
    properties = {
      storagePresentationContexts = {
        maxProperties = 127,
        minProperties = 1,
        patternProperties = {
          "^(?=.{1,64}$)(0|[1-9][0-9]*)(\\.(0|[1-9][0-9]*))*$" = {
            items = {
              pattern = "^(?=.{1,64}$)(0|[1-9][0-9]*)(\\.(0|[1-9][0-9]*))*$",
              type    = "string"
            },
            minItems    = 1,
            type        = "array",
            uniqueItems = true
          }
        },
        type = "object"
      },
    },
    type = "object"
  }
}

resource "echostream_managed_node_type" "test" {
  config_template = jsonencode(local.mnt_config)
  description     = "An HL7 inbounder Node definition"
  name            = "test"
  image_uri       = "226390263822.dkr.ecr.us-east-1.amazonaws.com/hl7-mllp-inbound-node:0.5-dev"
  mount_requirements = [
    {
      description = "my mount1"
      target      = "/foo"
    }
  ]
  port_requirements = [
    {
      container_port = 2575
      description    = "HL7 MLLP"
      protocol       = "tcp"
    }
  ]
}
