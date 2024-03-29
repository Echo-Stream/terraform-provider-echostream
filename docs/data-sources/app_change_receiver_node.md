---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "echostream_app_change_receiver_node Data Source - terraform-provider-echostream"
subcategory: ""
description: |-
  AppChangeReceiverNodes receive change messages from the AppChangeRouterNode. One per App, created when the App is created.
---

# echostream_app_change_receiver_node (Data Source)

AppChangeReceiverNodes receive change messages from the AppChangeRouterNode. One per App, created when the App is created.

## Example Usage

```terraform
resource "echostream_external_app" "test" {
  name = "test"
}

data "echostream_app_change_receiver_node" "test_receiver" {
  app = echostream_external_app.test.name
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app` (String) The App for this AppChangeReceiverNode

### Read-Only

- `description` (String) A human-readable description.
- `name` (String) The name of the Node. Must be unique within the Tenant.
- `receive_message_type` (String) The MessageType that this Node is capable of receiving.
