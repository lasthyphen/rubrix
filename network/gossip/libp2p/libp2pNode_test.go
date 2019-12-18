package libp2p

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"

	golog "github.com/ipfs/go-log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gologging "github.com/whyrusleeping/go-logging"
)

func TestLibP2PNode_Start_Stop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nodes, err := createLibP2PNodes(ctx, t, 1)
	assert.NoError(t, err)
	assert.NoError(t, nodes[0].Stop())
}

// TestLibP2PNode_GetPeerInfo checks that given a node name, the corresponding node id is consistently generated
// e.g. Node name: "node1" always generates the libp2p node id QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N
func TestLibP2PNode_GetPeerInfo(t *testing.T) {
	var nodes []NodeAddress
	var ps []peer.AddrInfo
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			nodes = append(nodes, NodeAddress{name: fmt.Sprintf("node%d", j), ip: "1.1.1.1", port: "0"})
			p, err := GetPeerInfo(nodes[j])
			require.NoError(t, err)
			if i == 0 {
				ps = append(ps, p)
			} else {
				assert.Equal(t, ps[j].ID.String(), p.ID.String(), fmt.Sprintf(" node ids not consistently generated"))
			}
		}
	}
}

// TestLibP2PNode_AddPeers checks if nodes can be added as peers to a given node
func TestLibP2PNode_AddPeers(t *testing.T) {
	t.Skip(" A libp2p issue causes this test to fail once in a while. Ignoring test")
	// A longer timeout is needed to overcome timeouts - https://github.com/ipfs/go-ipfs/issues/5800
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	// count value of 10 runs into this issue on localhost https://github.com/libp2p/go-libp2p-pubsub/issues/96
	// since localhost connection have short deadlines
	var count = 3
	// Create nodes
	nodes, err := createLibP2PNodes(ctx, t, count)
	require.NoError(t, err)
	defer func() {
		if nodes != nil {
			for _, n := range nodes {
				n.Stop()
			}
		}
	}()
	var ids []NodeAddress
	// Get actual ip and port numbers on which the nodes were started
	for _, n := range nodes[1:] {
		ip, p := n.GetIPPort()
		ids = append(ids, NodeAddress{name: n.name, ip: ip, port: p})
	}
	// To the 1st node add the remaining 9 nodes as peers.
	require.NoError(t, nodes[0].AddPeers(ctx, ids))
	actual := nodes[0].libP2PHost.Peerstore().Peers().Len()
	// Check if all 9 nodes have been added as peers to the first node.
	assert.Equal(t, count, actual, "peers expected: %d, found: %d", count, actual)

	// Check if libp2p reports node 1 is connected to the other nodes
	for _, a := range nodes[0].libP2PHost.Peerstore().Peers() {
		// A node is also a peer to itself but not marked as connected, hence skip checking that.
		if nodes[0].libP2PHost.ID().String() == a.String() {
			continue
		}
		assert.Eventuallyf(t, func() bool {
			return network.Connected == nodes[0].libP2PHost.Network().Connectedness(a)
		}, 3*time.Second, time.Millisecond, fmt.Sprintf(" node 0 not connected with %s", a.String()))
	}
}

// TestLibP2PNode_PubSub checks if nodes can subscribe to a topic and send and receive a message.
func TestLibP2PNode_PubSub(t *testing.T) {
	t.Skip(" A libp2p issue causes this test to fail once in a while. Ignoring test")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var count = 5
	golog.SetAllLoggers(gologging.INFO)

	// Step 1: Create nodes
	nodes, err := createLibP2PNodes(ctx, t, count)
	require.NoError(t, err)
	defer func() {
		if nodes != nil {
			for _, n := range nodes {
				n.Stop()
			}
		}
	}()

	// Step 2: Subscribe to a flow topic
	// A node will receive it's own message (https://github.com/libp2p/go-libp2p-pubsub/issues/65)
	// hence expect count and not count - 1 messages to be received (one by each node, including the sender)
	ch := make(chan string, count)
	for _, n := range nodes {
		m := n.name
		// callback that is registered
		cb := func(msg []byte) {
			assert.Equal(t, []byte("hello"), msg)
			ch <- m
		}
		require.NoError(t, n.Subscribe(ctx, Consensus, cb))
	}

	// Step 3: Connect a node to it's neighbour (Daisy chain them (1->2, 2->3...9->10))
	for i := 0; i < count-1; i++ {
		s := nodes[i]
		d := nodes[i+1]
		dip, dport := d.GetIPPort()
		nd := &NodeAddress{name: d.name, ip: dip, port: dport}
		require.NoError(t, s.AddPeers(ctx, []NodeAddress{*nd}))
		assert.Eventuallyf(t, func() bool {
			return network.Connected == s.libP2PHost.Network().Connectedness(d.libP2PHost.ID())
		}, 3*time.Second, time.Millisecond, fmt.Sprintf(" %s not connected with %s", s.name, d.name))
		e := 2
		if i%count == 0 {
			e = 1
		}
		assert.Equal(t, e, len(s.ps.ListPeers(string(Consensus))))
	}

	// Step 4: Wait for nodes to hearbeat
	time.Sleep(2 * time.Second)

	// Step 5: Publish a message from the first node and verify all nodes get it.
	// All nodes including node 0 - the sender, should receive it
	nodes[0].Publish(ctx, Consensus, []byte("hello"))
	recv := make(map[string]bool, count)
	for i := 0; i < count; i++ {
		select {
		case res := <-ch:
			recv[res] = true
		case <-time.After(10 * time.Second):
			missing := make([]string, 0)
			for _, n := range nodes {
				if _, found := recv[n.name]; !found {
					missing = append(missing, n.name)
				}
			}
			assert.Fail(t, " messages not received by nodes: "+strings.Join(missing, ", "))
			break
		}
	}

	// Step 6: Unsubscribe all nodes from the topic
	for _, n := range nodes {
		assert.NoError(t, n.UnSubscribe(Consensus))
	}
}

func createLibP2PNodes(ctx context.Context, t *testing.T, count int) (nodes []*P2PNode, err error) {
	defer func() {
		if err != nil && nodes != nil {
			for _, n := range nodes {
				n.Stop()
			}
		}
	}()
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	for i := 1; i <= count; i++ {
		var n = &P2PNode{}
		var nodeID = NodeAddress{name: fmt.Sprintf("node%d", i), ip: "0.0.0.0", port: "0"}
		err := n.Start(ctx, nodeID, l)
		require.NoError(t, err)
		require.Eventuallyf(t, func() bool {
			ip, p := n.GetIPPort()
			return ip != "" && p != ""
		}, 3*time.Second, time.Millisecond, fmt.Sprintf("node%d didn't start", i))
		nodes = append(nodes, n)
	}
	return nodes, err
}