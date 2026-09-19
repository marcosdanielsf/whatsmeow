package whatsmeow

import (
	"testing"

	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

func createNodeTags(req ReqCreateGroup) []string {
	var tags []string
	for _, node := range createGroupPrefixNodes(req) {
		tags = append(tags, node.Tag)
	}
	for _, node := range createGroupSuffixNodes(req) {
		tags = append(tags, node.Tag)
	}
	return tags
}

func hasTag(tags []string, want string) bool {
	for _, tag := range tags {
		if tag == want {
			return true
		}
	}
	return false
}

// A community create IQ that carries group-only settings is answered by WhatsApp with
// 400 bad-request, so IsParent must produce the <parent> node and nothing else.
func TestCreateGroupNodesCommunityOnlyCarriesParent(t *testing.T) {
	tags := createNodeTags(ReqCreateGroup{Name: "community", GroupParent: types.GroupParent{IsParent: true}})
	if len(tags) != 1 || tags[0] != "parent" {
		t.Fatalf("community create nodes = %v, want exactly [parent]", tags)
	}
}

func TestCreateGroupNodesCommunityKeepsDefaultApprovalMode(t *testing.T) {
	nodes := createGroupSuffixNodes(ReqCreateGroup{Name: "community", GroupParent: types.GroupParent{IsParent: true}})
	if len(nodes) != 1 {
		t.Fatalf("community suffix nodes = %d, want 1", len(nodes))
	}
	if got := nodes[0].Attrs["default_membership_approval_mode"]; got != "request_required" {
		t.Fatalf("default_membership_approval_mode = %v, want request_required", got)
	}
}

// Plain groups must keep every node they had before communities were fixed.
func TestCreateGroupNodesPlainGroupUnchanged(t *testing.T) {
	tags := createNodeTags(ReqCreateGroup{Name: "group"})
	for _, want := range []string{"member_add_mode", "ephemeral", "membership_approval_mode"} {
		if !hasTag(tags, want) {
			t.Fatalf("plain group create nodes = %v, want %s present", tags, want)
		}
	}
}

func TestCreateGroupNodesPlainGroupInsideCommunityKeepsLinkedParent(t *testing.T) {
	parent := types.JID{User: "120363318033230875", Server: types.GroupServer}
	nodes := createGroupSuffixNodes(ReqCreateGroup{Name: "group", GroupLinkedParent: types.GroupLinkedParent{LinkedParentJID: parent}})
	var linked *waBinary.Node
	for i := range nodes {
		if nodes[i].Tag == "linked_parent" {
			linked = &nodes[i]
		}
	}
	if linked == nil {
		t.Fatal("linked_parent node is missing for a group created inside a community")
	}
	if got := linked.Attrs["jid"]; got != parent {
		t.Fatalf("linked_parent jid = %v, want %v", got, parent)
	}
}
