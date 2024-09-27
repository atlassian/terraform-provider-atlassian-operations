# Jira Service Management Operations (JSM Ops) Terraform Provider

This project aims to enable users to manipulate Operations resources in Jira Service Management, via Terraform.
It is a functional replication of the _now transitioned_ [Opsgenie Provider](https://github.com/opsgenie/terraform-provider-opsgenie).

The provider is still under development. It currently supports the following resources:

* User\*
* Team
* Schedule (incl. Rotation)

\*Due to the internal structure of JSM, _user_ is implemented solely as a data source and supports **read operations only**.

### Related Links

- [Terraform Website](https://www.terraform.io)
- [Jira Service Management](https://www.atlassian.com/software/jira/service-management?tab=it-operations)
- [JSM Ops REST API](https://developer.atlassian.com/cloud/jira/service-desk-ops/rest/v2/intro/)

## How To Run Locally

The process to run the provider in a local environment requires the following steps:

1. Check the [Requirements](#1-requirements) section and install the missing components
2. Clone this repository
3. Compile & install the provider binary
4. Set local overrides, so terraform uses your local version of the provider
5. Test the configuration with a sample .tf file

### 1. Requirements

This project requires the following programs to be installed on your computer, and their main executables to be available in your PATH:

-	[Go](https://golang.org/doc/install) 1.22.7 (or higher, to build the provider plugin)
-	[Terraform](https://www.terraform.io/downloads.html) 1.9.5 (To test the plugin)


### 2. Cloning the repository

```bash
git clone git@bitbucket.org:jira-service-management/terraform-provider-jsm-ops.git
```

### 3. Compiling & Installing
_Make sure that go is already installed and the command is available on your path.
Check the [Go Documentation](https://go.dev/wiki/SettingGOPATH) for instructions on how to add the 
go executable to your path, if not already added._

While in the project directory, run the following commands:

```bash
go mod download
go install .
```
Go, _unless specified otherwise with the -o flag_, will install the resulting binary in the `$GOPATH/bin` directory.
If you do not know where your `$GOPATH` is, you can run:
```bash
go env GOPATH
```

### 4. Setting local overrides
This step is required for Terraform to use the plugin executable that you just compiled, 
instead of the one downloaded from the Terraform Registry.

* For macOS & Linux: Create a file called `.terraformrc` in your `$HOME` directory
* For Windows: Create a file called `terraform.rc` in your `%APPDATA%` directory

Add the following content to the file:

```hcl
provider_installation {
  
  # "/Users/<YOUR_USERNAME>/go/bin" is the path to the compiled provider ($GOPATH/bin).
  # Change it accordingly if your configuration is different.
   dev_overrides {
      # Replace <YOUR_USERNAME> with your username
      "registry.terraform.io/atlassian/jsm-ops" = "/Users/<YOUR_USERNAME>/go/bin"
   }

   # For all other providers, install them directly from their origin provider
   # registries as normal. If you omit this, Terraform will _only_ use
   # the dev_overrides block, and so no other providers will be available.
   direct {}
}
```

### 5. Test the configuration

* Create a basic terraform project locally with a `main.tf` file:

   ```hcl
   terraform {
      required_providers {
         jsm-ops = {
            source = "registry.terraform.io/atlassian/jsm-ops"
         }
      }
   }
   
   provider "jsm-ops" {
      cloud_id = "<YOUR_CLOUD_ID>"
      domain_name="<YOUR_DOMAIN>"      // e.g. domain.atlassian.net
      username = "<YOUR_USERNAME>"     // e.g. user@example.com
      password = "<YOUR_TOKEN_HERE>"   // e.g. API token created in Atlassian account settings
   }
   
   data "jsm-ops_user" "example" {
      email_address = "user1@example.com"
   }
   
   output "example" {
      value = "data.jsm-ops_user.example"
   }
   ```

* Run the following commands to test the provider:

   ```bash
   terraform plan
   terraform apply
   ```
   _`terraform init` is not necessary_