---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "echostream_cross_tenant_sending_node Resource - terraform-provider-echostream"
subcategory: ""
description: |-
  CrossTenantSendingNodes https://docs.echo.stream/docs/cross-tenant-sending-node send messages to a receiving Tenant.
---

# echostream_cross_tenant_sending_node (Resource)

[CrossTenantSendingNodes](https://docs.echo.stream/docs/cross-tenant-sending-node) send messages to a receiving Tenant.

## Example Usage

```terraform
data "echostream_processor_function" "json_2_xml" {
  name = "echo.json:echo.xml"
}

data "echostream_message_type" "json" {
  name = "echo.json"
}

resource "echostream_cross_tenant_sending_app" "test" {
  name             = "sending_app"
  receiving_app    = "receiving_app"
  receiving_tenant = "receiving_tenant"
}

resource "echostream_cross_tenant_sending_node" "test" {
  app                   = echostream_cross_tenant_sending_app.test.name
  description           = "my sending node"
  managed_processor     = data.echostream_processor_function.json_2_xml.name
  name                  = "test"
  receive_message_type  = data.echostream_message_type.json.name
  send_message_type     = "echo.xml"
  sequential_processing = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app` (String) The CrossTenantSendingApp this Node is associated with.
- `name` (String) The name of the Node. Must be unique within the Tenant.
- `receive_message_type` (String) The MessageType that this Node is capable of receiving.

### Optional

- `config` (String, Sensitive) The config, in JSON object format (i.e. - dict, map).
- `description` (String) A human-readable description.
- `inline_processor` (String) A Python code string that contains a single top-level function definition.This function is used as a template when creating custom processing in ProcessorNodesthat use this MessageType. This function must have the signature`(*, context, message, source, **kwargs)` and return None, a string or a list of strings. Mutually exclusive with `managedProcessor`.
- `logging_level` (String) The logging level. One of `DEBUG`, `ERROR`, `INFO`, `WARNING`. Defaults to `INFO`.
- `managed_processor` (String) The managedProcessor. Mutually exclusive with the `inlineProcessor`.
- `requirements` (Set of String) The list of Python requirements, in [pip](https://pip.pypa.io/en/stable/reference/requirement-specifiers/) format.
- `send_message_type` (String) The MessageType that this Node is capable of sending.
- `sequential_processing` (Boolean) `true` if messages should not be processed concurrently. If `false`, messages are processed concurrently. Defaults to `false`.

## Import

Import is supported using the following syntax:

```shell
terraform import echostream_cross_tenant_sending_node.sending "node_name"
```