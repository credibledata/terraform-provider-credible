package client

import "fmt"

// GroupMember represents a member of a group.
type GroupMember struct {
	UserGroupID string `json:"userGroupId,omitempty"`
	Status      string `json:"status,omitempty"`
}

// GroupMembersRequest is the request body for adding/removing group members.
type GroupMembersRequest struct {
	Members []GroupMember `json:"members"`
}

// GroupMembersResponse is the response body for listing group members.
type GroupMembersResponse struct {
	Members []GroupMember `json:"members"`
}

func (c *Client) ListGroupMembers(org, groupName string) ([]GroupMember, error) {
	var result GroupMembersResponse
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/groups/%s/members", org, groupName), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing group members for %q: %w", groupName, err)
	}
	return result.Members, nil
}

func (c *Client) AddGroupMembers(org, groupName string, members []GroupMember) error {
	body := GroupMembersRequest{Members: members}
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/groups/%s/members", org, groupName), body, nil)
	if err != nil {
		return fmt.Errorf("adding members to group %q: %w", groupName, err)
	}
	return nil
}

func (c *Client) RemoveGroupMembers(org, groupName string, members []GroupMember) error {
	body := GroupMembersRequest{Members: members}
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/groups/%s/members", org, groupName), body, nil)
	if err != nil {
		return fmt.Errorf("removing members from group %q: %w", groupName, err)
	}
	return nil
}

func (c *Client) UpdateGroupMemberStatus(org, groupName, memberName, status string) error {
	body := map[string]string{"status": status}
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/groups/%s/members/%s", org, groupName, memberName), body, nil)
	if err != nil {
		return fmt.Errorf("updating member %q status in group %q: %w", memberName, groupName, err)
	}
	return nil
}
