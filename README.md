# Atlassian Operations Terraform Provider

This project aims to enable users to manipulate Operations resources in Atlassian(Jira Service Management), via Terraform.
It is a functional replication of the _now transitioned_ [Opsgenie Provider](https://github.com/opsgenie/terraform-provider-opsgenie).

The provider is still under development. It currently supports the following resources:

* Team
* Schedule (**incl.** Rotation)
* Escalation
* Email Integration
* API-Based Integration

And the following data sources:
* User\*
* Team
* Schedule (**excl.** Rotation)

\*Due to the internal structure of Jira Service Management, _user_ is implemented solely as a data source and supports **read operations only**.

### Related Links

- [Terraform Website](https://www.terraform.io)
- [Jira Service Management](https://www.atlassian.com/software/jira/service-management?tab=it-operations)
- [JSM Ops REST API](https://developer.atlassian.com/cloud/jira/service-desk-ops/rest/v2/intro/)

## How To Run Locally

The process to run the provider in a local environment requires the following steps:

1. [Check the Requirements and install the missing components](#1-requirements)
2. [Clone this repository](#2-cloning-the-repository)
3. [Compile & install the provider binary](#3-compiling--installing)
4. [Set local overrides, so terraform uses your local version of the provider](#4-setting-local-overrides)
5. [Debugging](#5-debugging)
   1. [Create a simple .tf file](#51-create-a-simple-maintf-file)
   2. [Enable debugging](#52-enable-debugging)
6. [Running Acceptance Tests](#6-running-acceptance-tests)

### 1. Requirements

This project requires the following programs to be installed on your computer, and their main executables to be
available in your PATH:

-	[Go](https://golang.org/doc/install) 1.22 (or higher, to build the provider plugin)
-	[Terraform](https://www.terraform.io/downloads.html) 1.9.5 (To test the plugin)


### 2. Cloning the repository

```bash
git clone git@github.com:atlassian/terraform-provider-atlassian-operations.git
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
      "registry.terraform.io/atlassian/atlassian-operations" = "/Users/<YOUR_USERNAME>/go/bin"
   }

   # For all other providers, install them directly from their origin provider
   # registries as normal. If you omit this, Terraform will _only_ use
   # the dev_overrides block, and so no other providers will be available.
   direct {}
}
```

### 5. Debugging

#### 5.1 Create a simple `main.tf` file:

   ```hcl
   terraform {
   required_providers {
      atlassian-operations = {
         source = "registry.terraform.io/atlassian/atlassian-operations"
      }
   }
}

provider "atlassian-operations" {
   cloud_id = "<YOUR_CLOUD_ID>"
   domain_name="<YOUR_DOMAIN>"      // e.g. domain.atlassian.net
   email_address = "<YOUR_EMAIL_ADDRESS>"     // e.g. user@example.com
   token = "<YOUR_TOKEN_HERE>"   // e.g. API token created in Atlassian account settings
}

data "atlassian-operations_user" "example" {
   email_address = "user1@example.com"
}

output "example" {
   value = "data.atlassian-operations_user.example"
}
   ```

Instead of providing values in the _provider_ block directly, you can also set the following environment variables:

```bash
export ATLASSIAN_OPS_CLOUD_ID=YOUR_CLOUD_ID
export ATLASSIAN_OPS_DOMAIN_NAME=YOUR_DOMAIN
export ATLASSIAN_OPS_API_EMAIL_ADDRESS=YOUR_EMAIL_ADDRESS
export ATLASSIAN_OPS_API_TOKEN=YOUR_TOKEN
```

_If you do not want to debug the provider with a debugger, and would like to simply execute the Terraform file you
just created, you can skip the next part and jump directly to [Running Without Debugging](#53-running-without-debugging)_

#### 5.2. Enable Debugging

To enable debugging for the provider and make it connect to Delve before carrying on with the execution of the
instructions in the .tf file, you need to set the `debug` flag to `true`, via altering the `flag.BoolVar` statement in
the `main.tf` file:

```go
package main
import "flag"
// ...

func main() {
   var debug bool

   flag.BoolVar(&debug, "debug", true, "set to true to run the provider with support for debuggers like delve")
   flag.Parse()

   // Rest of the main.go file
}
```
This will make the provider executable pause the execution and wait for a debugger to connect before proceeding.

With the default configuration, the provider binary is compiled **without debug information**.
To include the necessary debug information within the library, you need to build the provider with the
-gcflags="all=-N -l" flag. Afterward, you can run and attach the process to the Delve debugger with the `dlv` command.

```bash
go build -gcflags="all=-N -l" .
dlv exec --accept-multiclient --continue --headless ./terraform_provider_jsm_ops -- -debug
```
_Most IDEs do building with debug flags and attaching to Delve debugger in the background automatically when the
debugger is run from within their UI._

When the provider executable is run with this configuration, it will output a message similar to the following:

```bash
API server listening at: 127.0.0.1:63446
debugserver-@(#)PROGRAM:LLDB  PROJECT:lldb-1600.0.36.3 for arm64.
Got a connection, launched process /Users/username/Library/Caches/JetBrains/GoLand2024.2/tmp/GoLand/___go_build_github_com_atlassian_terraform_provider_jsm_ops (pid = 47822).
{"@level":"debug","@message":"plugin address","@timestamp":"2024-10-02T00:03:48.057576+03:00","address":"/var/folders/5n/wcvl0l8d4nx15qz3jy9jn7wh0000gn/T/plugin3023012805","network":"unix"}
Provider started. To attach Terraform CLI, set the TF_REATTACH_PROVIDERS environment variable with the following:

        TF_REATTACH_PROVIDERS='{"registry.terraform.io/atlassian/atlassian-operations":{"Protocol":"grpc","ProtocolVersion":6,"Pid":47822,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/5n/wcvl0l8d4nx15qz3jy9jn7wh0000gn/T/plugin3023012805"}}}'
```

Simply follow the instructions as they are prompted. Either set the `TF_REATTACH_PROVIDERS` environment variable in
your terminal, or prepend it to your every Terraform command.

* Set the environment variable:
```bash
export TF_REATTACH_PROVIDERS=<_PROMPTED_STRING_AT_THE_DEBUG_CONSOLE_>
```

* Run the Terraform commands directly:
```bash
TF_REATTACH_PROVIDERS=<_PROMPTED_STRING_AT_THE_DEBUG_CONSOLE_> terraform plan
TF_REATTACH_PROVIDERS=<_PROMPTED_STRING_AT_THE_DEBUG_CONSOLE_> terraform apply
```

More information on how to use Delve with Terraform can be found in the
[Terraform documentation](https://developer.hashicorp.com/terraform/plugin/debugging).

### 6. Running Acceptance Tests
To run the acceptance tests, additional to the ones specified in the [Debugging](#51-create-a-simple-maintf-file) section, you need to set
the following environment variables as well:

```bash
export ATLASSIAN_ACCTEST_EMAIL_PRIMARY=USER_EMAIL
export ATLASSIAN_ACCTEST_EMAIL_SECONDARY=ANOTHER_USER_EMAIL
export ATLASSIAN_ACCTEST_ORGANIZATION_ID=ORGANIZATION_ID
export TF_ACC=1
```

Acceptance tests do not require a main.tf file to be present, as they are run directly from the test files. To run the acceptance tests,
simply run the following commands:

```bash
cd internal/provider
go test -count=1 -v
```

**Keep in mind that running acceptance tests will work on your existing JSM instance, which can result in notification emails being sent and extra usage fees.**