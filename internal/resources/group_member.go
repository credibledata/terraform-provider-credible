package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GroupMemberResource{}
var _ resource.ResourceWithImportState = &GroupMemberResource{}

type GroupMemberResource struct {
	client *client.Client
}

type GroupMemberResourceModel struct {
	Organization types.String `tfsdk:"organization"`
	GroupName    types.String `tfsdk:"group_name"`
	UserGroupID  types.String `tfsdk:"user_group_id"`
	Status       types.String `tfsdk:"status"`
}

func NewGroupMemberResource() resource.Resource {
	return &GroupMemberResource{}
}

func (r *GroupMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_member"
}

func (r *GroupMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages membership of a user or group within a Credible group.",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: "The organization name. Defaults to the provider's organization.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_name": schema.StringAttribute{
				Description: "The name of the group to add the member to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_group_id": schema.StringAttribute{
				Description: "The user or group identifier. Format: 'user:{email}' or 'group:{groupName}'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The membership role: 'admin' or 'member'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("admin", "member"),
				},
			},
		},
	}
}

func (r *GroupMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *GroupMemberResource) getOrg(model *GroupMemberResourceModel) string {
	if !model.Organization.IsNull() && !model.Organization.IsUnknown() {
		return model.Organization.ValueString()
	}
	return r.client.Organization
}

func (r *GroupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	if org == "" {
		resp.Diagnostics.AddError("Missing organization", "Organization must be set either on the resource or provider.")
		return
	}

	members := []client.GroupMember{
		{
			UserGroupID: plan.UserGroupID.ValueString(),
			Status:      plan.Status.ValueString(),
		},
	}

	err := r.client.AddGroupMembers(org, plan.GroupName.ValueString(), members)
	if err != nil {
		resp.Diagnostics.AddError("Error adding group member", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	members, err := r.client.ListGroupMembers(org, state.GroupName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading group members", err.Error())
		return
	}

	// Find this specific member in the list
	targetID := state.UserGroupID.ValueString()
	for _, m := range members {
		if m.UserGroupID == targetID {
			state.Organization = types.StringValue(org)
			state.Status = types.StringValue(m.Status)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	// Member not found — remove from state
	resp.State.RemoveResource(ctx)
}

func (r *GroupMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GroupMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)

	// The PATCH endpoint uses the member's email/name (the part after user:/group:)
	userGroupID := plan.UserGroupID.ValueString()
	parts := strings.SplitN(userGroupID, ":", 2)
	memberName := userGroupID
	if len(parts) == 2 {
		memberName = parts[1]
	}

	err := r.client.UpdateGroupMemberStatus(org, plan.GroupName.ValueString(), memberName, plan.Status.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating group member status", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	members := []client.GroupMember{
		{UserGroupID: state.UserGroupID.ValueString()},
	}

	err := r.client.RemoveGroupMembers(org, state.GroupName.ValueString(), members)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error removing group member", err.Error())
	}
}

func (r *GroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: org/groupName/user:email or org/groupName/group:otherGroup
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format: organization/group_name/user_group_id (e.g., my-org/my-group/user:alice@example.com)")
		return
	}

	org, groupName, userGroupID := parts[0], parts[1], parts[2]

	// Verify the member exists by listing members
	members, err := r.client.ListGroupMembers(org, groupName)
	if err != nil {
		resp.Diagnostics.AddError("Error importing group member", err.Error())
		return
	}

	for _, m := range members {
		if m.UserGroupID == userGroupID {
			state := GroupMemberResourceModel{
				Organization: types.StringValue(org),
				GroupName:    types.StringValue(groupName),
				UserGroupID:  types.StringValue(m.UserGroupID),
				Status:       types.StringValue(m.Status),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("Member not found", fmt.Sprintf("Member %q not found in group %q", userGroupID, groupName))
}
