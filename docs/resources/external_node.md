---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "echostream_external_node Resource - terraform-provider-echostream"
subcategory: ""
description: |-
  ExternalNodes https://docs.echo.stream/docs/external-node exist outside the EchoStream Cloud. Can be part of an ExternalApp or CrossAccountApp. You may use any computing resource or language that you want to implement them.
---

# echostream_external_node (Resource)

[ExternalNodes](https://docs.echo.stream/docs/external-node) exist outside the EchoStream Cloud. Can be part of an ExternalApp or CrossAccountApp. You may use any computing resource or language that you want to implement them.

## Example Usage

```terraform
data "echostream_message_type" "json" {
  name = "echo.json"
}

resource "echostream_external_app" "test" {
  name = "test"
}

resource "echostream_external_node" "test" {
  app                  = echostream_external_app.test.name
  name                 = "test"
  receive_message_type = data.echostream_message_type.json.name
  send_message_type    = data.echostream_message_type.json.name
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app` (String) The ExternalApp or CrossAccountApp this Node is associated with.
- `name` (String) The name of the Node. Must be unique within the Tenant.

### Optional

- `config` (String, Sensitive) The config, in JSON object format (i.e. - dict, map).
- `description` (String) A human-readable description.
- `receive_message_type` (String) The MessageType that this Node is capable of receiving.
- `send_message_type` (String) The MessageType that this Node is capable of sending.

## Import

Import is supported using the following syntax:

```shell
terraform import echostream_external_node.external "node_name"
```
