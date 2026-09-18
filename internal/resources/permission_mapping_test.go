package resources

import (
	"testing"

	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Plain unit tests -- no TF_ACC, no live API -- so they run in CI where the
// acceptance tests in this package self-skip.
//
// They pin the one thing that makes a permission manageable at all: the subject
// in state is the one the caller asked for, not the one the response happened to
// carry. Taking it from the response empties the id whenever the API omits the
// field, and every verb then addresses `.../permissions/` with no id.

// A permission response carrying the role and no subject.
func subjectlessResult(permission string) *client.Permission {
	return &client.Permission{Permission: permission}
}

func TestEnvironmentPermissionApplyResult_KeepsTheCallersSubject(t *testing.T) {
	model := EnvironmentPermissionResourceModel{
		Environment: types.StringValue("test-env"),
		UserGroupID: types.StringValue("group:ci-publisher"),
		Permission:  types.StringValue("viewer"),
	}

	model.applyResult("example-org", model.UserGroupID.ValueString(), subjectlessResult("modeler"))

	if got := model.UserGroupID.ValueString(); got != "group:ci-publisher" {
		t.Errorf("user_group_id: expected %q, got %q", "group:ci-publisher", got)
	}
	if got := model.Permission.ValueString(); got != "modeler" {
		t.Errorf("permission: expected %q, got %q", "modeler", got)
	}
	if got := model.Organization.ValueString(); got != "example-org" {
		t.Errorf("organization: expected %q, got %q", "example-org", got)
	}
	if got := model.Environment.ValueString(); got != "test-env" {
		t.Errorf("environment: expected %q, got %q", "test-env", got)
	}
}

// Import has no prior state to carry the subject, so it comes from the import ID.
func TestEnvironmentPermissionApplyResult_TakesTheSubjectFromTheImportID(t *testing.T) {
	var model EnvironmentPermissionResourceModel

	model.applyResult("example-org", "user:testuser@example.com", subjectlessResult("admin"))

	if got := model.UserGroupID.ValueString(); got != "user:testuser@example.com" {
		t.Errorf("user_group_id: expected %q, got %q", "user:testuser@example.com", got)
	}
}

// A response that names a different subject is not authority either: the path
// decided which grant was read or written.
func TestEnvironmentPermissionApplyResult_IgnoresASubjectInTheResponse(t *testing.T) {
	model := EnvironmentPermissionResourceModel{
		UserGroupID: types.StringValue("group:ci-publisher"),
	}

	model.applyResult("example-org", model.UserGroupID.ValueString(), &client.Permission{
		UserGroupID: "user:someone-else@example.com",
		Permission:  "viewer",
	})

	if got := model.UserGroupID.ValueString(); got != "group:ci-publisher" {
		t.Errorf("user_group_id: expected %q, got %q", "group:ci-publisher", got)
	}
}

func TestOrganizationPermissionApplyResult_KeepsTheCallersSubject(t *testing.T) {
	model := OrganizationPermissionResourceModel{
		UserGroupID: types.StringValue("group:ci-publisher"),
		Permission:  types.StringValue("member"),
	}

	model.applyResult("example-org", model.UserGroupID.ValueString(), subjectlessResult("admin"))

	if got := model.UserGroupID.ValueString(); got != "group:ci-publisher" {
		t.Errorf("user_group_id: expected %q, got %q", "group:ci-publisher", got)
	}
	if got := model.Permission.ValueString(); got != "admin" {
		t.Errorf("permission: expected %q, got %q", "admin", got)
	}
	if got := model.Organization.ValueString(); got != "example-org" {
		t.Errorf("organization: expected %q, got %q", "example-org", got)
	}
}

func TestOrganizationPermissionApplyResult_TakesTheSubjectFromTheImportID(t *testing.T) {
	var model OrganizationPermissionResourceModel

	model.applyResult("example-org", "user:testuser@example.com", subjectlessResult("admin"))

	if got := model.UserGroupID.ValueString(); got != "user:testuser@example.com" {
		t.Errorf("user_group_id: expected %q, got %q", "user:testuser@example.com", got)
	}
}
