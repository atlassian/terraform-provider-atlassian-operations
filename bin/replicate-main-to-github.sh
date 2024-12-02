#!/usr/bin/env bash

git remote rename origin bitbucket
git remote add origin git@github.com:atlassian/terraform-provider-atlassian-operations.git
git push origin master