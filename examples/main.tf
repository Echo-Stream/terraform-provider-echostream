terraform {
  required_providers {
    echostream = {
      source = "echo.stream/dev/echostream"
    }
  }
}

provider "echostream" {
  appsync_endpoint = "https://api-dev.us-east-1.echo.stream/graphql"
  client_id        = "30pseo6e4rmn4q89nttqkerhbp"
  password         = "DV*?S;]Oa0q*tE7)<;()Xf%XPV=QuHO9mR2XeZa8}Ra91El]y#OArPDGU*bZMAu;b*lPxaKU8|}MH)+CtnNF]CFf?OEc~^i;9w}+^e:<9^)%Hq)UTQ.M0eV&QKkCXos,.UBIIzH|<J[%P=1t1V,X06dLZ=[OKS(^mE=i:>*&7fbs[eVo|d:x,jbP{XX2K<qzO8uG+mv]K+g9Pz)9;DI#CeazCcICt?k[y6q(R#KgGc}zQ=hVAU#kGE|..^&v*hmG"
  tenant           = "Foobar3000"
  username         = "e4fc580be92b47ca91c5e98fa1c665df"
  user_pool_id     = "us-east-1_qlGAO4m1H"
}

data "echostream_alert_emitter_node" "aln" {
}

data "echostream_app_change_router_node" "acrn" {

}

data "echostream_audit_emitter_node" "aun" {
}

data "echostream_change_emitter_node" "cen" {
}

data "echostream_tenant" "tenant" {

}

data "echostream_message_type" "text" {
  name = "echo.text"
}

data "echostream_api_authenticator_function" "basic" {
  name = "echo.basic"
}

data "echostream_processor_function" "csv_2_json" {
  name = "echo.csv:echo.json"
}

locals {
  tenant_config = {
    value_1 = "my value"
    value_2 = 1.2
  }
}

resource "echostream_tenant" "tenant" {
  description = "This is my description redux"
  config      = jsonencode(local.tenant_config)
}

resource "echostream_message_type" "test" {
  auditor            = "def auditor(*, message, **kwargs):\n    print(\"foo\")\n    return {}\n"
  bitmapper_template = "def bitmapper(*, context, message, source, **kwargs):\n\n    from decimal import Decimal\n    import simplejson as json\n\n    message = json.loads(message, parse_float=Decimal)\n\n    bitmap = 0x0\n\n    # TODO - Perform conditional bitmapping of the JSON message here.\n    # The returned bitmap will be compared against the routes in the\n    # route table for route matching. This will be done by\n    # (message_bitmap \u0026 route) == route. If no routes match, the\n    # message will be filtered.\n\n    return bitmap\n"
  description        = "Test message type"
  name               = "test2"
  processor_template = "def processor(*, context, message, source, **kwargs):\n\n    from decimal import Decimal\n    import simplejson as json\n\n    message = json.loads(message, parse_float=Decimal)\n\n    # TODO - Perform any transformations to the message here.\n    # This can include transforming the message to something that\n    # is no longer JSON. Remember, you MUST return a string or None.\n    # If None is returned, the message will be filtered.\n\n    return json.dumps(message, separators=(\",\", \":\"), use_decimal=True)\n"
  readme             = "# echo.json\n\nRepresents a [JavaScript Object Notation (JSON)](https://en.wikipedia.org/wiki/JSON) formatted message.\n\nJSON is extensively used to transport data, provide input and results from API's, and store data in many different databases. It is very software language friendly, with almost all modern languages supporting it inherently.\n\nAn example of JSON formatted data:\n```json\n{\n    \"firstName\": \"John\",\n    \"lastName\": \"Smith\",\n    \"isAlive\": true,\n    \"age\": 27,\n    \"address\": {\n    \"streetAddress\": \"21 2nd Street\",\n    \"city\": \"New York\",\n    \"state\": \"NY\",\n    \"postalCode\": \"10021-3100\"\n    },\n    \"phoneNumbers\": [\n        {\n            \"type\": \"home\",\n            \"number\": \"212 555-1234\"\n        },\n        {\n            \"type\": \"office\",\n            \"number\": \"646 555-4567\"\n        }\n    ],\n    \"children\": [],\n    \"spouse\": null\n}\n```\n\nThe only requirement for this message type is correctly formatted JSON."
  sample_message     = "{\n  \"firstName\": \"John\",\n  \"lastName\": \"Smith\",\n  \"isAlive\": true,\n  \"age\": 27,\n  \"address\": {\n    \"streetAddress\": \"21 2nd Street\",\n    \"city\": \"New York\",\n    \"state\": \"NY\",\n    \"postalCode\": \"10021-3100\"\n  },\n  \"phoneNumbers\": [\n    {\n      \"type\": \"home\",\n      \"number\": \"212 555-1234\"\n    },\n    {\n      \"type\": \"office\",\n      \"number\": \"646 555-4567\"\n    }\n  ],\n  \"children\": [],\n  \"spouse\": null\n}"
}
