---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "echostream_change_emitter_node Data Source - terraform-provider-echostream"
subcategory: ""
description: |-
  ChangeEmitterNodes https://docs.echo.stream/docs/change-emitter-node emit change messages. One per Tenant, automatically created when the Tenant is created.
---

# echostream_change_emitter_node (Data Source)

[ChangeEmitterNodes](https://docs.echo.stream/docs/change-emitter-node) emit change messages. One per Tenant, automatically created when the Tenant is created.

## Example Usage

```terraform
data "echostream_change_emitter_node" "test" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `description` (String) A human-readable description.
- `name` (String) The name of the Node. Must be unique within the Tenant.
- `send_message_type` (String) The MessageType that this Node is capable of sending.
