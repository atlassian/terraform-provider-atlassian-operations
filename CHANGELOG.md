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