package main

import (
	"context"
	"fmt"
)

// inviteMember invites a user as an enterprise member.
func (c *client) inviteMember(ctx context.Context, enterpriseID, login string) error {
	const m = `mutation($ent:ID!,$login:String!){
      inviteEnterpriseMember(input:{enterpriseId:$ent, invitee:$login}){ clientMutationId }
    }`
	return c.do(ctx, m, map[string]any{"ent": enterpriseID, "login": login}, nil)
}

// inviteAdmin invites a user as an enterprise owner (admin).
func (c *client) inviteAdmin(ctx context.Context, enterpriseID, login string) error {
	const m = `mutation($ent:ID!,$login:String!){
      inviteEnterpriseAdmin(input:{enterpriseId:$ent, invitee:$login, role:OWNER}){ clientMutationId }
    }`
	return c.do(ctx, m, map[string]any{"ent": enterpriseID, "login": login}, nil)
}

// setAdminRole changes an existing administrator's role (OWNER or BILLING_MANAGER).
func (c *client) setAdminRole(ctx context.Context, enterpriseID, login, role string) error {
	const m = `mutation($ent:ID!,$login:String!,$role:EnterpriseAdministratorRole!){
      updateEnterpriseAdministratorRole(input:{enterpriseId:$ent, login:$login, role:$role}){ clientMutationId }
    }`
	return c.do(ctx, m, map[string]any{"ent": enterpriseID, "login": login, "role": role}, nil)
}

// removeAdmin removes a user's enterprise owner role entirely.
func (c *client) removeAdmin(ctx context.Context, enterpriseID, login string) error {
	const m = `mutation($ent:ID!,$login:String!){
      removeEnterpriseAdmin(input:{enterpriseId:$ent, login:$login}){ clientMutationId }
    }`
	return c.do(ctx, m, map[string]any{"ent": enterpriseID, "login": login}, nil)
}

// removeMember removes a user from the enterprise (and all its orgs).
func (c *client) removeMember(ctx context.Context, enterpriseID, login string) error {
	// removeEnterpriseMember takes a user ID, so resolve the login first.
	uid, err := c.userID(ctx, login)
	if err != nil {
		return fmt.Errorf("resolve user %q: %w", login, err)
	}
	const m = `mutation($ent:ID!,$user:ID!){
      removeEnterpriseMember(input:{enterpriseId:$ent, userId:$user}){ clientMutationId }
    }`
	return c.do(ctx, m, map[string]any{"ent": enterpriseID, "user": uid}, nil)
}

func (c *client) userID(ctx context.Context, login string) (string, error) {
	const q = `query($login:String!){ user(login:$login){ id } }`
	var out struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := c.do(ctx, q, map[string]any{"login": login}, &out); err != nil {
		return "", err
	}
	if out.User.ID == "" {
		return "", fmt.Errorf("user %q not found", login)
	}
	return out.User.ID, nil
}
