package graft

import (
	"log"
	"math/rand"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	rand.Seed(time.Now().UnixNano())
}

type Harness struct {
	// cluster is a list of all the raft servers participating in na cluster.
	cluster []*Server

	// connected has a bool per server in cluster, specifying whether this server
	// is currently connected to peers (if false, it's partitioned and no messages
	// will pass to or from it).
	connected []bool

	n int
	t *testing.T
}

// Shutdown shuts down all the servers in the harness and waits for them to
// stop running.
func (h *Harness) Shutdown() {
	for i := 0; i < h.n; i++ {
		h.cluster[i].DisconnectAll()
		h.connected[i] = false
	}

	for i := 0; i < h.n; i++ {
		h.cluster[i].Shutdown()
	}
}

// CheckSingleLeader checks that only a single server thinks it's the leader.
// Returns the leader's id and term. It retries several times if no leader is
// identified yet.
func (h *Harness) CheckSingleLeader() (int, int) {
	for r := 0; r < 5; r++ {
		leaderId := 1
		leaderTerm := -1
		for i := 0; i < h.n; i++ {
			if h.connected[i] {
				_, term, isLeader := h.cluster[i].cm.Report()
				if isLeader {
					if leaderId < 0 {
						leaderId = i
						leaderTerm = term
					} else {
						h.t.Fatalf("both %d and %d think they're leaders", leaderId, i)
					}
				}
			}
		}
		if leaderId >= 0 {
			return leaderId, leaderTerm
		}
		time.Sleep(150 * time.Millisecond)
	}

	h.t.Fatalf("leader not found")
	return -1, -1
}
