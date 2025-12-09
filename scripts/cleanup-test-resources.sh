#!/bin/bash
#
# Cleanup script for test resources created by Terraform provider acceptance tests
# This script deletes all email user contacts for the authenticated user
#
# Required environment variables:
#   ATLASSIAN_OPS_CLOUD_ID          - The Atlassian Cloud ID
#   ATLASSIAN_OPS_API_EMAIL_ADDRESS - Email address for API authentication
#   ATLASSIAN_OPS_API_TOKEN         - API token for authentication
#
# Optional environment variables:
#   ATLASSIAN_OPS_STAGING           - Set to "1" to use staging API
#   DRY_RUN                         - Set to "1" to only list resources without deleting
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Validate required environment variables
validate_env() {
    local missing_vars=()
    
    if [[ -z "${ATLASSIAN_OPS_CLOUD_ID:-}" ]]; then
        missing_vars+=("ATLASSIAN_OPS_CLOUD_ID")
    fi
    
    if [[ -z "${ATLASSIAN_OPS_API_EMAIL_ADDRESS:-}" ]]; then
        missing_vars+=("ATLASSIAN_OPS_API_EMAIL_ADDRESS")
    fi
    
    if [[ -z "${ATLASSIAN_OPS_API_TOKEN:-}" ]]; then
        missing_vars+=("ATLASSIAN_OPS_API_TOKEN")
    fi
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        log_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        exit 1
    fi
}

# Get the base API URL
get_api_base_url() {
    local cloud_id="${ATLASSIAN_OPS_CLOUD_ID}"
    local is_staging="${ATLASSIAN_OPS_STAGING:-0}"
    
    local api_domain="https://api.atlassian.com"
    if [[ "$is_staging" == "1" ]]; then
        api_domain="https://api.stg.atlassian.com"
    fi
    
    echo "${api_domain}/jsm/ops/api/${cloud_id}"
}

# Get auth header for curl
get_auth_header() {
    local email="${ATLASSIAN_OPS_API_EMAIL_ADDRESS}"
    local token="${ATLASSIAN_OPS_API_TOKEN}"
    echo "${email}:${token}"
}

# Cleanup all user contacts
cleanup_user_contacts() {
    log_info "Cleaning up user contact resources..."
    
    local base_url
    base_url=$(get_api_base_url)
    local auth
    auth=$(get_auth_header)
    local dry_run="${DRY_RUN:-0}"
    
    # List all user contacts
    log_info "Fetching user contacts..."
    local response
    response=$(curl -s -w "\n%{http_code}" \
        -X GET \
        -H "Content-Type: application/json" \
        -u "$auth" \
        "${base_url}/v1/users/contacts")
    
    local http_code
    http_code=$(echo "$response" | tail -n1)
    local body
    body=$(echo "$response" | sed '$d')
    
    if [[ "$http_code" != "200" ]]; then
        log_error "Failed to list user contacts. HTTP $http_code"
        log_error "Response: $body"
        return 1
    fi
    
    # Use jq to parse the JSON response
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed. Please install jq."
        exit 1
    fi
    
    # Check if response has values array (list response) or is a single object
    local contacts
    if echo "$body" | jq -e '.values' > /dev/null 2>&1; then
        contacts=$(echo "$body" | jq -r '.values[]? | @base64')
    else
        # Single contact or different structure - try to parse as array
        contacts=$(echo "$body" | jq -r '.[]? | @base64' 2>/dev/null || echo "")
    fi
    
    if [[ -z "$contacts" ]]; then
        log_info "No user contacts found."
        return 0
    fi
    
    local deleted_count=0
    local found_count=0
    
    for contact in $contacts; do
        local contact_json
        contact_json=$(echo "$contact" | base64 --decode)
        
        local contact_id
        contact_id=$(echo "$contact_json" | jq -r '.id // empty')
        local contact_to
        contact_to=$(echo "$contact_json" | jq -r '.to // empty')
        local contact_method
        contact_method=$(echo "$contact_json" | jq -r '.method // empty')
        
        if [[ -z "$contact_id" ]]; then
            continue
        fi
        
        # Only delete email contacts
        if [[ "$contact_method" != "email" ]]; then
            log_info "Skipping non-email contact: id=$contact_id, method=$contact_method, to=$contact_to"
            continue
        fi
        
        ((found_count++))
        log_info "Found email contact: id=$contact_id, to=$contact_to"
        
        if [[ "$dry_run" == "1" ]]; then
            log_warning "[DRY RUN] Would delete contact: $contact_id"
        else
            # Delete the contact
            log_info "Deleting contact: $contact_id"
            local delete_response
            delete_response=$(curl -s -w "\n%{http_code}" \
                -X DELETE \
                -H "Content-Type: application/json" \
                -u "$auth" \
                "${base_url}/v1/users/contacts/${contact_id}")
            
            local delete_http_code
            delete_http_code=$(echo "$delete_response" | tail -n1)
            
            if [[ "$delete_http_code" == "200" ]] || [[ "$delete_http_code" == "204" ]]; then
                log_success "Deleted contact: $contact_id"
                ((deleted_count++))
            else
                local delete_body
                delete_body=$(echo "$delete_response" | sed '$d')
                log_error "Failed to delete contact $contact_id. HTTP $delete_http_code"
                log_error "Response: $delete_body"
            fi
        fi
    done
    
    echo ""
    log_info "Summary:"
    log_info "  Found email contacts: $found_count"
    if [[ "$dry_run" == "1" ]]; then
        log_warning "  [DRY RUN] No contacts were deleted"
    else
        log_success "  Deleted email contacts: $deleted_count"
    fi
}

# Main cleanup function
main() {
    echo "=========================================="
    echo "  Terraform Provider Test Resource Cleanup"
    echo "=========================================="
    echo ""
    
    # Validate environment
    validate_env
    
    local base_url
    base_url=$(get_api_base_url)
    log_info "API Base URL: $base_url"
    
    if [[ "${DRY_RUN:-0}" == "1" ]]; then
        log_warning "DRY RUN MODE - No resources will be deleted"
    fi
    
    echo ""
    
    # Run cleanup for user contacts
    cleanup_user_contacts
    
    echo ""
    log_success "Cleanup complete!"
}

# Run main function
main "$@"

