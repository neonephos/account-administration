package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// action is one planned change against the enterprise.
type action struct {
	kind  string // "invite-owner", "invite-member", "promote", "demote", "remove-owner", "remove-member"
	login string
	desc  string
}

// plan computes the actions needed to move `live` toward `cfg`.
type reconcilePlan struct {
	actions  []action
	adds     int // owner/member invites
	removes  int // owner/member removals (for the removal-delta guard)
	liveSize int
}

func computePlan(cfg *Config, live *liveState) reconcilePlan {
	wantOwners := toSet(cfg.Spec.Owners)
	wantMembers := toSet(cfg.Spec.Members)

	var p reconcilePlan
	p.liveSize = len(live.owners) + len(live.members)

	// Desired owners.
	for _, login := range cfg.Spec.Owners {
		l := strings.ToLower(login)
		switch {
		case live.owners[l] != "":
			// already an owner — nothing to do
		case live.members[l] != "":
			p.actions = append(p.actions, action{"promote", login, "promote member -> owner"})
		default:
			p.actions = append(p.actions, action{"invite-owner", login, "invite as owner"})
			p.adds++
		}
	}

	// Desired members.
	for _, login := range cfg.Spec.Members {
		l := strings.ToLower(login)
		switch {
		case live.members[l] != "":
			// already a member — nothing to do
		case live.owners[l] != "":
			p.actions = append(p.actions, action{"demote", login, "demote owner -> member"})
		default:
			p.actions = append(p.actions, action{"invite-member", login, "invite as member"})
			p.adds++
		}
	}

	// Live owners not desired anywhere -> remove owner role.
	for l, orig := range live.owners {
		if !wantOwners[l] && !wantMembers[l] {
			p.actions = append(p.actions, action{"remove-owner", orig, "remove enterprise owner"})
			p.removes++
		}
	}
	// Live members not desired anywhere -> remove from enterprise.
	for l, orig := range live.members {
		if !wantOwners[l] && !wantMembers[l] {
			p.actions = append(p.actions, action{"remove-member", orig, "remove from enterprise"})
			p.removes++
		}
	}

	sort.SliceStable(p.actions, func(i, j int) bool {
		if p.actions[i].kind != p.actions[j].kind {
			return p.actions[i].kind < p.actions[j].kind
		}
		return strings.ToLower(p.actions[i].login) < strings.ToLower(p.actions[j].login)
	})
	return p
}

// safety checks before applying.
type guards struct {
	minOwners       int
	maxRemovalDelta float64
}

func (p reconcilePlan) check(cfg *Config, g guards) error {
	if err := checkMinOwners(cfg, g); err != nil {
		return err
	}
	if p.liveSize > 0 && g.maxRemovalDelta < 1.0 {
		ratio := float64(p.removes) / float64(p.liveSize)
		if ratio > g.maxRemovalDelta {
			return fmt.Errorf("plan removes %d of %d enterprise people (%.2f), exceeding --max-removal-delta=%.2f (typo guard)",
				p.removes, p.liveSize, ratio, g.maxRemovalDelta)
		}
	}
	return nil
}

// checkMinOwners is a config-only guard that can run before any network call.
func checkMinOwners(cfg *Config, g guards) error {
	if n := len(cfg.Spec.Owners); n < g.minOwners {
		return fmt.Errorf("config declares %d owner(s), which is below --min-owners=%d (lockout guard)", n, g.minOwners)
	}
	return nil
}

// apply executes the plan. When confirm is false it only prints (dry-run).
func (p reconcilePlan) apply(ctx context.Context, c *client, enterpriseID string, confirm bool) error {
	if len(p.actions) == 0 {
		fmt.Println("no changes: enterprise owners/members already match config")
		return nil
	}
	for _, a := range p.actions {
		if !confirm {
			fmt.Printf("DRY-RUN would %-14s %s (%s)\n", a.kind, a.login, a.desc)
			continue
		}
		fmt.Printf("%-14s %s (%s)\n", a.kind, a.login, a.desc)
		var err error
		switch a.kind {
		case "invite-owner":
			err = c.inviteAdmin(ctx, enterpriseID, a.login)
		case "invite-member":
			err = c.inviteMember(ctx, enterpriseID, a.login)
		case "promote":
			err = c.setAdminRole(ctx, enterpriseID, a.login, "OWNER")
		case "demote":
			// Demote to member = remove admin role, then ensure membership. Removing
			// the admin role keeps the user as an enterprise member.
			err = c.removeAdmin(ctx, enterpriseID, a.login)
		case "remove-owner":
			err = c.removeAdmin(ctx, enterpriseID, a.login)
		case "remove-member":
			err = c.removeMember(ctx, enterpriseID, a.login)
		default:
			err = fmt.Errorf("unknown action %q", a.kind)
		}
		if err != nil {
			return fmt.Errorf("%s %s: %w", a.kind, a.login, err)
		}
	}
	return nil
}
