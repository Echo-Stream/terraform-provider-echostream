---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "echostream_external_app Resource - terraform-provider-echostream"
subcategory: ""
description: |-
  ExternalApps https://docs.echo.stream/docs/external-app provide a way to process messages in their Nodes using any compute resource.
---

# echostream_external_app (Resource)

[ExternalApps](https://docs.echo.stream/docs/external-app) provide a way to process messages in their Nodes using any compute resource.

## Example Usage

```terraform
resource "echostream_external_app" "ex_app" {
  name = "ex_app"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the app; must be unique in the Tenant.

### Optional

- `config` (String, Sensitive) The config for the app. All nodes in the app will be allowed to access this. Must be a JSON object.
- `description` (String) A human-readable description of the app.
- `table_access` (Boolean) Indicates if this app can gain access to the Tenant's DynamoDB [table](https://docs.echo.stream/docs/table).

### Read-Only

- `appsync_endpoint` (String) The EchoStream AppSync Endpoint that this ExternalApp must use.
- `audit_records_endpoint` (String) The app-specific endpoint for posting audit records. Details about this endpoint may be found [here](https://docs.echo.stream/docs/auditing-messages-from-cross-accountexternalmanaged-apps#auditing-without-use-of-the-echostreamnode-package).
- `credentials` (Attributes) The AWS Cognito Credentials that allow the app to access the EchoStream GraphQL API. (see [below for nested schema](#nestedatt--credentials))

<a id="nestedatt--credentials"></a>
### Nested Schema for `credentials`

Read-Only:

- `client_id` (String) The AWS Cognito Client ID used to connect to EchoStream.
- `password` (String, Sensitive) The password to use when connecting to EchoStream.
- `user_pool_id` (String) The AWS Cognito User Pool ID used to connect to EchoStream.
- `username` (String) The username to use when connecting to EchoStream.

## Import

Import is supported using the following syntax:

```shell
terraform import echostream_external_app.xapp "app_name"
```