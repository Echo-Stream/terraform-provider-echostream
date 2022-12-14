---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "echostream_managed_node Resource - terraform-provider-echostream"
subcategory: ""
description: |-
  ManagedNodes https://docs.echo.stream/docs/managed-node are instances of Docker containers that exist within ManagedApps.
---

# echostream_managed_node (Resource)

[ManagedNodes](https://docs.echo.stream/docs/managed-node) are instances of Docker containers that exist within ManagedApps.

## Example Usage

```terraform
data "echostream_managed_node_type" "hl7_in" {
  name = "echo.hl7-mllp-inbound-node"
}

resource "echostream_managed_app" "test" {
  name = "test"
}

resource "echostream_managed_node" "test" {
  app               = echostream_managed_app.test.name
  managed_node_type = data.echostream_managed_node_type.hl7_in.name
  name              = "test"
  ports = [
    {
      container_port = 2575
      host_port      = 2575
      protocol       = "tcp"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app` (String) The ManagedApp that this Node is associated with.
- `managed_node_type` (String) The ManagedNodeType of this ManagedNode. This Node must conform to all of the config, mount and port requirements specified in the ManagedNodeType.
- `name` (String) The name of the Node. Must be unique within the Tenant.

### Optional

- `config` (String, Sensitive) The config, in JSON object format (i.e. - dict, map).
- `description` (String) A human-readable description.
- `logging_level` (String) The logging level. One of `DEBUG`, `ERROR`, `INFO`, `WARNING`. Defaults to `INFO`.
- `mounts` (Attributes Set) A list of the mounts (i.e. - volumes) used by the Docker container. (see [below for nested schema](#nestedatt--mounts))
- `ports` (Attributes Set) A list of ports exposed by the Docker container. (see [below for nested schema](#nestedatt--ports))

### Read-Only

- `receive_message_type` (String) The MessageType that this Node is capable of receiving.
- `send_message_type` (String) The MessageType that this Node is capable of sending.

<a id="nestedatt--mounts"></a>
### Nested Schema for `mounts`

Required:

- `target` (String) The path to mount the volume in the Docker container.

Optional:

- `source` (String) The source of the mount. If not present, an anonymous volume will be created.

Read-Only:

- `description` (String) A human-readable description.


<a id="nestedatt--ports"></a>
### Nested Schema for `ports`

Required:

- `container_port` (Number) The exposed container port.
- `host_port` (Number) The exposed host port. Must be between `1024` and `65535`, inclusive.
- `protocol` (String) The protocol to use for the port. One of `sctp`, `tcp` or `udp`.

Optional:

- `host_address` (String) The host address the port is exposed on. Defaults to `0.0.0.0`.

Read-Only:

- `description` (String) A human-readable description.

## Import

Import is supported using the following syntax:

```shell
terraform import echostream_managed_node.managed "node_name"
```
