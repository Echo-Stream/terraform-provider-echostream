---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "echostream_tenant_user Resource - terraform-provider-echostream"
subcategory: ""
description: |-
  TenantUsers https://docs.echo.stream/docs/users-1 are used to interact with your Tenant via the UI.
---

# echostream_tenant_user (Resource)

[TenantUsers](https://docs.echo.stream/docs/users-1) are used to interact with your Tenant via the UI.

## Example Usage

```terraform
resource "echostream_tenant_user" "john_doe" {
  email = "john.doe@mail.com"
  role  = "admin"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) The user's email address.
- `role` (String) The ApiUser's role. Must be one of `admin`, `read_only`, or `user`.

### Optional

- `status` (String) The status. If set, must be one of `active` or `inactive`.

### Read-Only

- `first_name` (String) The user's first name, if available.
- `last_name` (String) The user's last name, if available.

## Import

Import is supported using the following syntax:

```shell
terraform import echostream_tenant_user.user "user@email.com"
```
