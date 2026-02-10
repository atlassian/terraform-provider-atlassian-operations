# Team can be imported by providing the team id and the organization id, separated by a comma
terraform import atlassian-operations_team.example "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

# When teams are scoped to a site, provide the site id as the third parameter
terraform import atlassian-operations_team.example "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
