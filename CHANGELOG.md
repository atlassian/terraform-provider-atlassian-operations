## v2.0.5

#### Product:

- Fixed schedule rotation restrictions order mismatch by normalizing the response

## v2.0.4

#### Product:

- Fixed a perpetual state drift bug on every alert policy update. The bug was causing Terraform to show redundant changes even though nothing had changed in the alert policy configuration.

## v2.0.3

#### Product:

- Added retry logic for `delete_default_resources`: if an empty list is received, Terraform will retry multiple times to fetch actions before discarding directly.
- Added more detailed logs for the `delete_default_resources` flow.

## v2.0.2

#### Product:

- Support optional `site_id` in team resource import for site-scoped teams.

## v2.0.1

#### Product:

- Added `siteId` parameter to the fetch team members API call, which was failing during team creation due to the missing parameter.

## v2.0.0

#### Product:

- Fixed an issue where policy order updates were failing silently, causing Terraform drift and potential resources to be created with incorrect order values.

## v1.1.10

#### Product:

- Added `StartDay` and `EndDay` configuration for the Notification Policy resource.

## v1.1.9

#### Resources:

Implemented Jira Service Management Service Resource
## v1.1.8

#### Product:

- Add `delete_default_actions` attribute in the API Integration resource. Fix Order attribute problem on Notification Policy and Alert Policy resources. Handle external deletion of resources by setting their state to Deleted on Read operations. This ensures that the provider can handle resources that have been deleted outside of Terraform, preventing state inconsistencies.
## v1.1.7

#### Product:

- Bug fix `delete_default_resources` attribute in the Team resource. List Schedules in paginated manner to find the default schedule and delete it if `delete_default_resources` is set to true. This ensures that the default schedule is removed when the team is created with this attribute set to true.
## v1.1.6

#### Product:

- Added support for the `delete_default_resources` attribute in the Team resource. This allows users to remove the default escalation and schedule for newly created teams. Be cautious, as this also changes the team routing rule to None, requiring you to define a routing rule separately.


## v1.1.5

#### Product:

- Fixed an issue where the "field" attribute of an Integration Action's Conditions attribute had an invalid validator, causing Terraform not to accept valid values.

## v1.1.4

#### Product:

- Fixed an issue where the api_key value was being removed from the state file (where it was previously present) after consecutive terraform apply runs.
  - This does not affect the api_key field in the resource, which is still only available after Create operations. Reading an existing integration will not return this field.

## v1.1.3

#### Product:

- Fixed an issue where the users were unable to create global alert policies.

## v1.1.2

#### Product:

- Fixed an issue where notification policy resources that don't specify the "not" parameter in the conditions clause caused an inconsistency error after applying. [JSDCLOUD-16983](https://jira.atlassian.com/browse/JSDCLOUD-16983)

## v1.1.1

#### Resources:

Implemented Integration Action Resource
Implemented Heartbeat Resource
Implemented Maintenance Resource
Implemented Notification Policy Resource

#### Product:
Added API Key field for API Integration Resource. **This field is only available after Create operations. Reading an existing integration will not return this field.**

## v1.1.0

#### Product:

- Added Compass Operations Support

#### Resources:

- Implemented Custom Role resource 
- Implemented Alert Policy resource 
- Implemented User Contact resource

## v1.0.3

#### Resources:

- Implemented Notification Rule Resource
- Implemented Routing Rule Resource
- Updated documentation for existing resources

## v1.0.2

### FIXES:

- [JSDCLOUD-16292](https://jira.atlassian.com/browse/JSDCLOUD-16292): Added a workaround for Date-Time Format Mismatch between OPS API and Terraform Provider
- Corrected wrong parameter name (username -> email_address) in the example provider configuration
- Provide a more descriptive error message an API request repeatedly fails, instead of only providing the retry count

## v1.0.1

### FIXES:

- Fixed a race condition issue due to consequent requests sent by the same HTTP client effecting each other.
- Fixed typos in schedule and schedule rotation import scripts

## v1.0.0

### FEATURES:

Initial Release

#### Resources:

- Implemented Team Resource
- Implemented Schedule Resource
- Implemented Schedule Rotation Resource
- Implemented Escalation Resource
- Implemented Integration Resources

#### Data Sources:

- Implemented Team Data Source
- Implemented Schedule Data Source
- Implemented User Data Source