package redis

import "testing"

func TestUpdateNodeClusterInfoFillsLoadingInstanceFromClusterNodes(t *testing.T) {
	clusterNodesInfo := [][]string{
		{"master-id", "127.0.0.1:6379", "0-5460", "master", "-"},
		{"slave-id", "127.0.0.1:6380", "", "slave", "master-id"},
	}

	master := &Instance{Addr: "127.0.0.1:6379", LoadingError: true}
	master.UpdateNodeClusterInfo(clusterNodesInfo)

	if master.NodeID != "master-id" {
		t.Fatalf("master NodeID = %q, want %q", master.NodeID, "master-id")
	}
	if master.Role != "master" {
		t.Fatalf("master Role = %q, want %q", master.Role, "master")
	}
	if got := master.GetSlotCount(); got != 5461 {
		t.Fatalf("master slot count = %d, want %d", got, 5461)
	}

	slave := &Instance{Addr: "127.0.0.1:6380", LoadingError: true}
	slave.UpdateNodeClusterInfo(clusterNodesInfo)

	if slave.NodeID != "slave-id" {
		t.Fatalf("slave NodeID = %q, want %q", slave.NodeID, "slave-id")
	}
	if slave.Role != "slave" {
		t.Fatalf("slave Role = %q, want %q", slave.Role, "slave")
	}
	if slave.Master != "127.0.0.1:6379" {
		t.Fatalf("slave Master = %q, want %q", slave.Master, "127.0.0.1:6379")
	}
}
