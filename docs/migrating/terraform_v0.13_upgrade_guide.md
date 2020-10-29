
# Terraform 0.13 Upgrade Guide
> Comprehensive guide to [0.13upgrade command](https://www.terraform.io/docs/commands/0.13upgrade.html).
> This command is available only in Terraform v0.13 releases


1. Ensure you have downloaded terraform v0.13 binary.  
1. Run *$> terraform 0.13upgrade* command.

   >By default, 0.13upgrade changes configuration files in the current working directory. However, you can provide an explicit path to another directory if desired, which may be useful for automating migrations of several modules in the same repository.
   >note: For batch usage, refer to [0.13upgrade command](https://www.terraform.io/docs/commands/0.13upgrade.html).
   
   A new file called version.tf will be created if terraform block does not already exists.
1. Modify version.tf to include source and version
    ```
    terraform {
        required_providers {
            ddcloud = {
                source = "github.local/DimensionDataResearch/ddcloud"
                version = "3.0.5"
            }
        }
        required_version = ">= 0.13"
    }
    ```

1. As ddcloud provider is register in terraform registry, thus, we can't simply use namespace to locate ddcloud provider. 
However, we could configure terraform to look for ddcloud provider locally, considered as in-house provider plugin in a specific location. 

- Create .terraformrc in user home directory. 
`$> touch ~/.terraformrc` 
> By default terraform looks for .terraformrc from home directory. Alternatively, you could use environment variable  `TF_CLI_CONFIG_FILE` to specify a custom location.

- Edit .terraformrc like below:
    ```
    provider_installation {
      filesystem_mirror {
        path    = "[replace with your HOME dir]/.terraform.d/plugins"
        include = ["github.local/*/*"]
      }
      direct {
        exclude = ["github.local/*/*"]
      }
    }
    ```
  >Note: This config tell terraform to look for the plugin locally rather than using online registry.
  
- Create plugin folder structure as described in .terraformrc and drop the provider plugin binary here 
  
  `[HOME_Directory]/.terraform.d/plugins/github.local/DimensionDataResearch/ddcloud/13.0.1/darwin_amd64/terraform-provider-ddcloud`
 > Note: I place the the provider plugin in darwin_amd64 folder as I'm using on mac. You may want to use 'linux_amd64' or  'windows_amd64' that matches
 > your OS environment.

       
