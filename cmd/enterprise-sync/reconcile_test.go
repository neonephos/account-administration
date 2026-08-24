package main

import "testing"

func liveFrom(owners, members []string) *liveState {
	st := &liveState{owners: map[string]string{}, members: map[string]string{}}
	for _, o := range owners {
		st.owners[lower(o)] = o
	}
	for _, m := range members {
		st.members[lower(m)] = m
	}
	return st
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

func kinds(p reconcilePlan) map[string]int {
	m := map[string]int{}
	for _, a := range p.actions {
		m[a.kind]++
	}
	return m
}

func TestComputePlan_NoOp(t *testing.T) {
	cfg := &Config{Enterprise: "e", Spec: Spec{Owners: []string{"alice", "bob"}, Members: []string{"carol"}}}
	live := liveFrom([]string{"alice", "bob"}, []string{"carol"})
	p := computePlan(cfg, live)
	if len(p.actions) != 0 {
		t.Fatalf("expected no actions, got %v", p.actions)
	}
}

func TestComputePlan_InviteAndRemove(t *testing.T) {
	cfg := &Config{Enterprise: "e", Spec: Spec{Owners: []string{"alice", "bob"}, Members: []string{"dave"}}}
	// live: alice owner (keep), zoe owner (remove), carol member (remove), bob missing (invite owner), dave missing (invite member)
	live := liveFrom([]string{"alice", "zoe"}, []string{"carol"})
	p := computePlan(cfg, live)
	k := kinds(p)
	if k["invite-owner"] != 1 || k["invite-member"] != 1 || k["remove-owner"] != 1 || k["remove-member"] != 1 {
		t.Fatalf("unexpected plan kinds: %v (actions=%v)", k, p.actions)
	}
	if p.removes != 2 {
		t.Fatalf("expected 2 removes, got %d", p.removes)
	}
}

func TestComputePlan_PromoteDemote(t *testing.T) {
	cfg := &Config{Enterprise: "e", Spec: Spec{Owners: []string{"alice", "carol"}, Members: []string{"bob"}}}
	// live: alice owner (keep), bob owner (demote), carol member (promote)
	live := liveFrom([]string{"alice", "bob"}, []string{"carol"})
	p := computePlan(cfg, live)
	k := kinds(p)
	if k["promote"] != 1 || k["demote"] != 1 {
		t.Fatalf("expected 1 promote + 1 demote, got %v (actions=%v)", k, p.actions)
	}
	if p.removes != 0 {
		t.Fatalf("promote/demote must not count as removals, got %d", p.removes)
	}
}

func TestComputePlan_CaseInsensitiveLogins(t *testing.T) {
	cfg := &Config{Enterprise: "e", Spec: Spec{Owners: []string{"Alice", "Bob"}}}
	live := liveFrom([]string{"alice", "bob"}, nil)
	p := computePlan(cfg, live)
	if len(p.actions) != 0 {
		t.Fatalf("case differences should be a no-op, got %v", p.actions)
	}
}

func TestCheck_MinOwners(t *testing.T) {
	cfg := &Config{Enterprise: "e", Spec: Spec{Owners: []string{"solo"}}}
	p := computePlan(cfg, liveFrom([]string{"solo"}, nil))
	if err := p.check(cfg, guards{minOwners: 2, maxRemovalDelta: 1.0}); err == nil {
		t.Fatal("expected min-owners guard to fail with a single owner")
	}
}

func TestCheck_RemovalDelta(t *testing.T) {
	cfg := &Config{Enterprise: "e", Spec: Spec{Owners: []string{"alice", "bob"}}}
	// live has 10 people; config keeps 0 of the members -> big removal.
	members := []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8"}
	live := liveFrom([]string{"alice", "bob"}, members)
	p := computePlan(cfg, live)
	if err := p.check(cfg, guards{minOwners: 2, maxRemovalDelta: 0.25}); err == nil {
		t.Fatal("expected removal-delta guard to trip when removing 8/10")
	}
	// With the guard relaxed to 1.0 it should pass.
	if err := p.check(cfg, guards{minOwners: 2, maxRemovalDelta: 1.0}); err != nil {
		t.Fatalf("relaxed guard should pass: %v", err)
	}
}
