package gossip

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/dapperlabs/flow-go/crypto"
	"github.com/dapperlabs/flow-go/proto/gossip/messages"
)

// extractHashMsgInfo receives the bytes of a HashMessage instance, unmarshals it and returns
// its components
func extractHashMsgInfo(hashMsgBytes []byte) ([]byte, *messages.Socket, error) {
	hashMsg := &messages.HashMessage{}
	if err := proto.Unmarshal(hashMsgBytes, hashMsg); err != nil {
		return nil, &messages.Socket{}, fmt.Errorf("could not unmarshal message: %v", err)
	}

	return hashMsg.GetHashBytes(), hashMsg.GetSenderSocket(), nil
}

// Pick a random subset of a list
func randomSubSet(list []string, size int) ([]string, error) {
	if size == 0 {
		return nil, fmt.Errorf("could not find subset of a size 0 ")
	}

	if len(list) < size {
		return list, nil
	}

	var (
		gossipPartners = make([]string, size)
		us             = newUniqSelector(len(list))
		index          = 0
		err            error
	)

	for i := 0; i < size; i++ {
		index, err = us.Int()
		if err != nil {
			return nil, fmt.Errorf("could not pick index for a gossip partner: %v", err)
		}
		gossipPartners[i] = list[index]
	}
	return gossipPartners, nil
}

// computeHash computes the hash of GossipMessage using sha256 algorithm
func computeHash(msg *messages.GossipMessage) ([]byte, error) {

	alg, err := crypto.NewHasher(crypto.SHA3_256)
	if err != nil {
		return nil, err
	}
	msgHash := alg.ComputeHash(msg.GetPayload())

	return msgHash, nil
}