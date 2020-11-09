# ddcloud\_server\_backup

This resource enables Cloud Backup for a server and, optionally, configures its backup clients.

## Example Usage

```hcl
resource "ddcloud_server_backup" "myserver" {
  server            = "${ddcloud_server.myserver.id}"
  service_plan      = "Essentials"

  client {
    type            = "FA.Linux"
    schedule_policy = "6AM - 12PM"
    storage_policy  = "14 Day Storage Policy"
  }
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) The Id of the server for which Cloud Backup will be enabled / configured.
* `service_plan` - (Required) The Cloud Backup service plan to use.  
  Must be one of `Essentials`, `Advanced`, or `Enterprise`.
* `client` - (Optional) A list of Backup clients to be assigned to / configured on the server.  
  **Note**: Once created, do not change the order of the backup clients in this list.
  * `type` - (Required) The type of backup client (e.g. `FA.Linux`).
  * `schedule_policy` - (Required) The name of the schedule policy to use (e.g. `6AM - 12PM`).
  * `storage_policy` - (Required) The name of the storage policy to use (e.g. `14 Day Storage Policy`).
  * `alert` - (Optional) The client's alerting configuration.  
    * `trigger` - (Required) The trigger for backup client alerts.  
      Must be one of `ON_FAILURE`, `ON_SUCCESS`, or `ON_SUCCESS_OR_FAILURE`.
    * `emails` - (Required) A list of one or more email addresses that alerts will be sent to.

## Attribute Reference

* `asset_id` - The asset Id assigned to the server by Cloud Backup.
* `client` - Computed attributes for the server's backup clients.
  * `client.id` - The Id assigned to the backup client by Cloud Backup.
  * `client.description` - A short description of the backup client type.
  * `client.download_url` - The URL where the backup client installer can be downloaded.
  * `client.status` - The client status (e.g. `Unregistered`, `Offline`, `Active`, etc).

## Import

Import of `ddcloud_server_backup` is not implemented yet.
