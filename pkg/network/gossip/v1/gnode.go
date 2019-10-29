package gnode

import (
	"context"
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/dapperlabs/flow-go/pkg/grpc/shared"
	"github.com/dapperlabs/flow-go/pkg/network/gossip"
	"github.com/dapperlabs/flow-go/pkg/network/gossip/v1/order"
	"github.com/rs/zerolog"
)

// To make sure that Node complies with the gossip.Service interface
var _ gossip.Service = (*Node)(nil)

// Node is holding the required information for a functioning async gossip node
type Node struct {
	logger  zerolog.Logger
	regMngr *registryManager
	queue   chan *order.Order

	hashCache hashCache
	msgStore  messageDatabase

	// QueueSize is the buffer size of the node for holding incoming Gossip Messages
	// Once buffer of a node gets full, it does not accept incoming messages
	queueSize int

	address         *shared.Socket
	peers           []string
	fanoutSet       []string
	staticFanoutNum int

	// server represents the protocol on which gnode will operate
	server ServePlacer
}

// NewNode returns a new instance of a Gossip node with a predefined registry of message types, a set of peers
// and a staticFanoutNum indicating the size of the static fanout
func NewNode(config *NodeConfig) *Node {
	node := &Node{
		logger:          config.logger,
		regMngr:         config.regMnger,
		queueSize:       config.queueSize,
		queue:           make(chan *order.Order, config.queueSize),
		hashCache:       newMemoryHashCache(),
		msgStore:        newMemMsgDatabase(),
		address:         config.address,
		peers:           config.peers,
		staticFanoutNum: config.staticFanoutNum,
	}
	node.initDefault()
	return node
}

// SetProtocol sets the underlying protocol for sending and receiving messages. The protocol should
// implement ServePlacer
func (n *Node) SetProtocol(s ServePlacer) {
	n.server = s
}

// syncHashProposal receives a hash of a message that is designated to be sent as SyncGossip() , so
// that in case it has not been received, the function returns the address of the current node
// prompting the sender to send the original message.
// HashMsgBytes: represents the hash of the message being proposed
// Todo: This function is planned to be upgraded to a streaming version so that no return value is required upon a negative answer
func (n *Node) syncHashProposal(ctx context.Context, hashMsgBytes []byte) ([]byte, error) {
	log := n.logger.With().
		Str("function", "syncHashProposal").
		Logger()

	// extract the information from the hash message
	hashBytes, _, err := extractHashMsgInfo(hashMsgBytes)
	if err != nil {
		err = errors.Wrap(err, "could not extract hash message info")
		log.Debug().Err(err).Send()
		return nil, err
	}

	// turn bytes of hash to a string
	hash := string(hashBytes)

	// if the message has been already received ignore the proposal
	if n.hashCache.isReceived(hash) {
		return nil, nil
	}

	// set the message as confirmed
	n.hashCache.confirm(hash)

	// turn the address to bytes
	address, err := proto.Marshal(n.address)
	if err != nil {
		log.Debug().Err(err).Send()
		return nil, errors.Wrap(err, "could not marshal address")
	}

	// return the address so that the sender sends the original message
	return address, nil
}

// asyncHashProposal receives a hash of a message that is designated to be sent as an AsyncGossip, so that
// if it has not been received yet, a "requestMessage" is sent to the sender prompting him to send the original message
// hashMsgBytes: represents the hash of the message being proposed.
func (n *Node) asyncHashProposal(ctx context.Context, hashMsgBytes []byte) ([]byte, error) {
	log := n.logger.With().
		Str("function", "asyncHashProposal").
		Logger()

	// extract the information from the HashMessage
	hashBytes, senderAddr, err := extractHashMsgInfo(hashMsgBytes)
	if err != nil {
		err = errors.Wrap(err, "could not extract hash message info")
		log.Debug().Err(err).Send()
		return nil, err
	}

	// turn bytes of hash to a string
	hash := string(hashBytes)

	// if the message has been already received ignore the proposal
	if n.hashCache.isReceived(hash) {
		return nil, nil
	}
	// set the message as confirmed
	n.hashCache.confirm(hash)

	// this message will request the original message
	requestMsg, err := generateHashMessage(hashBytes, n.address)
	if err != nil {
		log.Debug().Err(err).Send()
		return nil, errors.Wrap(err, "could not generate hash message")
	}

	requestMsgBytes, err := proto.Marshal(requestMsg)
	if err != nil {
		log.Debug().Err(err).Send()
		return nil, errors.Wrap(err, "could not marshal message")
	}

	id, err := n.regMngr.MsgTypeToID("requestMessage")
	if err != nil {
		log.Debug().Err(err).Send()
		return nil, errors.Wrap(err, "could not get message type id")
	}
	//slice with the address to send the message to
	address := []string{socketToString(senderAddr)}

	// send a requestMessage to the sender to make him send the original message
	_, err = n.gossip(context.Background(), requestMsgBytes, address, id, false, gossip.ModeOneToOne)
	if err != nil {
		log.Debug().Err(err).Send()
		return nil, errors.Wrap(err, "could not gossip message")
	}

	return nil, nil
}

// requestMessage receives a hash of a message with the address of its sender, so that if the message
// exists in the store, it is sent to the requester.
// hashMsgBytes: represents the hash of the requested message.
func (n *Node) requestMessage(ctx context.Context, hashMsgBytes []byte) ([]byte, error) {
	log := n.logger.With().
		Str("function", "requestMessage").
		Logger()

	// extract information from hash message
	hashBytes, senderAddr, err := extractHashMsgInfo(hashMsgBytes)
	if err != nil {
		err = errors.Wrap(err, "could not extract hash message info")
		log.Debug().Err(err).Send()
		return nil, err
	}

	// turn bytes of hash into a string
	hash := string(hashBytes)

	// if the requested message does exist return an error
	if !n.hashCache.isReceived(hash) {
		err = fmt.Errorf("no message of given hash is not found")
		log.Debug().Err(err).Send()
		return nil, fmt.Errorf("no message of given hash is not found")
	}

	// obtain message from store
	msg, err := n.msgStore.Get(hash)
	if err != nil {
		log.Debug().Err(err).Send()
		return nil, errors.Wrap(err, "could not get message from database")
	}

	//slice with the address to send the message to
	address := []string{socketToString(senderAddr)}

	// send the original message
	_, err = n.gossip(ctx, msg.GetPayload(), address, msg.GetMessageType(), false, gossip.ModeOneToOne)
	if err != nil {
		log.Debug().Err(err).Send()
		return nil, errors.Wrap(err, "could not gossip message")
	}

	return nil, nil
}

// SyncGossip synchronizes over the reply of recipients
// i.e., it sends a message to all recipients and blocks for their reply
// payload:    represents the payload of the message
// recipients: the list of nodes designated to recieve the message
// msgType:    represents the type of message to be gossiped
func (n *Node) SyncGossip(ctx context.Context, payload []byte, recipients []string, msgType string) ([]*shared.GossipReply, error) {
	log := n.logger.With().
		Str("function", "SyncGossip").
		Logger()

	id, err := n.regMngr.MsgTypeToID(msgType)
	if err != nil {
		log.Debug().Err(err).Send()
		return nil, errors.Wrap(err, "could not get message type ID")
	}

	// generate GossipMessage type in order to hash it
	msg, err := generateGossipMessage(payload, recipients, id)
	if err != nil {
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, errors.Wrap(err, "could not generate gossip message")
	}
	_, err = n.tryStore(msg)
	// hash the GossipMessage
	hash, err := computeHash(msg)
	if err != nil {
		err = errors.Wrap(err, "could not generate hash")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	// generate the HashMessage to be sent in the hash proposal
	hashMsg, err := generateHashMessage(hash, n.address)
	if err != nil {
		err = errors.Wrap(err, "could not generate hash message")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	// marshal HashMessage to bytes
	hashMsgBytes, err := proto.Marshal(hashMsg)
	if err != nil {
		err = errors.Wrap(err, "could not marshal message")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	// realRecipients will store the addresses of nodes who want the actual message
	realRecipients := make([]string, 0)

	msgID, err := n.regMngr.MsgTypeToID("syncHashProposal")
	if err != nil {
		err = errors.Wrap(err, "could not get message type ID")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	var mode gossip.Mode
	switch len(recipients) {
	case 0:
		mode = gossip.ModeOneToAll
	case 1:
		mode = gossip.ModeOneToOne
	default:
		mode = gossip.ModeOneToMany
	}

	// send syncHashProposal
	replies, err := n.gossip(context.Background(), hashMsgBytes, recipients, msgID, true, mode)

	if err != nil {
		err = errors.Wrap(err, "could not gossip message: %v")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	// iterate over replies, and for non-nil replies, they must be addresses
	// so add them to realRecipients
	for _, rep := range replies {
		if rep.ResponseByte != nil {
			socket := &shared.Socket{}
			err := proto.Unmarshal(rep.ResponseByte, socket)

			if err != nil {
				// TODO: discuss handling erroneous address returns, we ignore them at the moment
				continue
			}

			realRecipients = append(realRecipients, socketToString(socket))
		}
	}
	// if no recipients, short circuit and return (Otherwise, if no one wants the message, OneToAll occurs)
	if len(realRecipients) == 0 {
		return []*shared.GossipReply{}, nil
	}

	// send actual message to addresses who request it
	return n.gossip(context.Background(), payload, realRecipients, id, true, mode)
}

// AsyncGossip synchronizes over the delivery
// i.e., sends a message to all recipients, and only blocks for delivery without blocking for their response
// payload:    represents the payload of the message
// recipients: the list of nodes designated to receive the message
// msgType:    represents the type of message to be gossiped (e.g., the engine type, the method type to be invoked over the payload, etc)
func (n *Node) AsyncGossip(ctx context.Context, payload []byte, recipients []string, msgType string) ([]*shared.GossipReply, error) {
	log := n.logger.With().
		Str("function", "AsyncGossip").
		Logger()

	msgID, err := n.regMngr.MsgTypeToID(msgType)
	if err != nil {
		err = errors.Wrap(err, "could not get message type ID")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	// generate GossipMessage type in order to hash it
	msg, err := generateGossipMessage(payload, recipients, msgID)
	if err != nil {
		err = errors.Wrap(err, "could not generate gossip message")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}
	_, err = n.tryStore(msg)

	// hash the GossipMessage
	hash, err := computeHash(msg)
	if err != nil {
		err = errors.Wrap(err, "could not generate hash")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	// generate the HashMessage to be sent in the hash proposal
	hashMsg, err := generateHashMessage(hash, n.address)
	if err != nil {
		err = errors.Wrap(err, "could not generate hash message")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	// marshal HashMessage to bytes
	hashMsgBytes, err := proto.Marshal(hashMsg)
	if err != nil {
		err = errors.Wrap(err, "could not marshal message")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	msgID, err = n.regMngr.MsgTypeToID("asyncHashProposal")
	if err != nil {
		err = errors.Wrap(err, "could not get message type ID")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	var mode gossip.Mode
	switch len(recipients) {
	case 0:
		mode = gossip.ModeOneToAll
	case 1:
		mode = gossip.ModeOneToOne
	default:
		mode = gossip.ModeOneToMany
	}

	// send the hash proposal to recipients
	_, err = n.gossip(ctx, hashMsgBytes, recipients, msgID, false, mode)
	if err != nil {
		err = errors.Wrap(err, "could not gossip message")
		log.Debug().Err(err).Send()
		return []*shared.GossipReply{}, err
	}

	return []*shared.GossipReply{}, nil
}

// AsyncQueue is invoked remotely using the gRPC stub,
// it receives a message from a remote node and places it inside the local nodes queue
// it is synchronized with the remote node on the message reception (and NOT reply), i.e., blocks the remote node until either
// a timeout or placement of the message into the queue
// msg: the message to be placed in the queue
func (n *Node) AsyncQueue(ctx context.Context, msg *shared.GossipMessage) (*shared.GossipReply, error) {

	log := n.logger.With().
		Str("function", "AsyncQueue").
		Logger()

	//if context timed out already, return an error
	select {
	case <-ctx.Done():
		log.Debug().Err(ErrTimedOut).Send()
		return &shared.GossipReply{}, ErrTimedOut
	default:
	}

	msgExists, err := n.tryStore(msg)

	if err != nil {
		err = errors.Wrap(err, "error in storing message")
		log.Debug().Err(err).Send()
		return &shared.GossipReply{}, err
	}

	// if message was confirmed, then ignore it
	if msgExists {
		return &shared.GossipReply{}, nil
	}

	// no recipient for a message means that the message is a gossip to ALL so the node also broadcasts that message
	switch len(msg.GetRecipients()) {
	// If the message is One To All, then forward it as well
	case 0:
		// In case the recipients list was empty, then this message is one
		// to all.
		go n.gossip(ctx, msg.GetPayload(), nil, msg.GetMessageType(), false, gossip.ModeOneToAll)
		// If the message is One To One, do nothing
	case 1:
		// Do nothing as this connection is one to one and we are the end point
	default:
		// In case the recipients list was greater than 1, then this message is one
		// to many.
		go n.gossip(ctx, msg.GetPayload(), msg.GetRecipients(), msg.GetMessageType(), false, gossip.ModeOneToMany)
	}

	// Wait until the context expires, or if the msg got placed on the queue
	select {
	case <-ctx.Done():
		log.Debug().Err(ErrTimedOut).Send()
		return &shared.GossipReply{}, ErrTimedOut
	case n.queue <- order.NewAsync(ctx, msg):
		return &shared.GossipReply{}, nil
	}
}

// SyncQueue is invoked remotely using the gRPC stub,
// it receives a message from a remote node and places it inside the local nodes queue
// it is synchronized with the remote node on the message reply, i.e., blocks the remote node until either
// a timeout or a reply is getting prepared
// msg: the message to be placed in the queue
func (n *Node) SyncQueue(ctx context.Context, msg *shared.GossipMessage) (*shared.GossipReply, error) {
	ord := order.NewSync(ctx, msg)

	log := n.logger.With().
		Str("function", "SyncQueue").
		Logger()

	//if context timed out already, return an error
	select {
	case <-ctx.Done():
		log.Debug().Err(ErrTimedOut).Send()
		return &shared.GossipReply{}, ErrTimedOut
	default:
	}

	// check if message is stored, and store it if it was not stored
	msgExists, err := n.tryStore(msg)
	if err != nil {
		err = errors.Wrap(err, "error in storing message")
		log.Debug().Err(err).Send()
		return &shared.GossipReply{}, err
	}
	// if the message was already confirmed, ignore it
	if msgExists {
		return &shared.GossipReply{}, nil
	}

	// extract the message id corresponding to the "syncHashProposal" message type to check the incoming message is
	// not a hash proposal if it is not a hash proposal then it will be gossiped to static fanout
	// this is done to avoid make a node only "gossip" upon having the actual message, and NOT only on receiving
	// a hash proposal
	msgID, err := n.regMngr.MsgTypeToID("syncHashProposal")
	if err != nil {
		log.Debug().Err(err).Send()
		return &shared.GossipReply{}, errors.Wrap(err, "could not get message type ID")
	}

	// if no recipients specified that means it is OneToAll, pass on the message to the static fanout.
	// if a message is a hash proposal, it should not be gossiped until the original message is captured
	if msg.MessageType != msgID {
		go func(ctx context.Context, msg *shared.GossipMessage) {
			msgType, err := n.regMngr.IDtoMsgType(msg.MessageType)
			if err != nil {
				log.Debug().Err(err).Send()
				return
			}
			// Gossip it to fanout
			// NOTE: the messages returned from this call are not used.
			switch len(msg.GetRecipients()) {
			// If the message is One To All, then forward it as well
			case 0:
				go n.SyncGossip(ctx, msg.GetPayload(), nil, msgType)
				// If the message is One To One, do nothing
			case 1:
				// Do nothing as this connection is one to one and we are the end point
			default:
				go n.SyncGossip(ctx, msg.GetPayload(), msg.GetRecipients(), msgType)
			}
		}(ctx, msg)
	}

	// getting blocked until either a timeout,
	// or the message getting placed into the local nodes queue
	select {
	case <-ctx.Done():
		log.Debug().Err(ErrTimedOut).Send()
		return &shared.GossipReply{}, ErrTimedOut
	case n.queue <- ord:
	}

	/*
		Part II: blocking on processing
	*/
	// getting blocked until either a timeout
	// or the message getting processed out of the queue of the local node
	select {
	case <-ctx.Done():
		log.Debug().Err(ErrTimedOut).Send()
		return &shared.GossipReply{}, ErrTimedOut
	case <-ord.Done():
	}
	msgReply, msgErr := ord.Result()
	if msgErr != nil {
		log.Debug().Err(msgErr).Send()
		return &shared.GossipReply{}, msgErr
	}
	return &shared.GossipReply{ResponseByte: msgReply}, msgErr
}

// Serve starts an async node grpc server, and its sweeper as well
func (n *Node) Serve(listener net.Listener) error {
	// Send of the sweeper to monitor the queue
	go n.sweeper()
	//TODO: make serve return error
	n.server.Serve(listener)

	return nil
}

// initDefault adds the default functions to the registry of the node
func (n *Node) initDefault() {
	// TODO: handle the error
	err := n.pickGossipPartners()
	if err != nil {
		n.logger.Debug().Err(err).Send()
	}
	n.RegisterFunc("syncHashProposal", n.syncHashProposal)
	n.RegisterFunc("asyncHashProposal", n.asyncHashProposal)
	n.RegisterFunc("requestMessage", n.requestMessage)
}

// RegisterFunc allows the addition of new message types to the node's registry
// msgType:   type of message to be registered
// handlFunc: the signature of the registered function
func (n *Node) RegisterFunc(msgType string, f HandleFunc) error {
	return n.regMngr.AddMessageType(msgType, f)
}

// gossip is the main function through which messages of all types are gossiped
// payload:       represents the payload of the function
// recipients:    the list of recipients of the gossip message, a nil represents a gossip to all the network
// msgType:       represents the type of the message being gossiped
// isSynchronous: the type of gossiping (Sync,Async)
func (n *Node) gossip(ctx context.Context, payload []byte, recipients []string, msgType uint64, isSynchronous bool, mode gossip.Mode) ([]*shared.GossipReply, error) {
	log := n.logger.With().
		Str("function", "gossip").
		Logger()

	// casting gossip message attributes into a protobuf message
	gossipMsg, err := generateGossipMessage(payload, recipients, msgType)
	if err != nil {
		err := errors.Wrap(err, "could not generate gossip message")
		log.Error().Err(err).Send()
		return nil, err
	}

	// make sure message is in storage before gossiping
	_, err = n.tryStore(gossipMsg)

	if err != nil {
		err := errors.Wrap(err, "error in storing message")
		log.Error().Err(err).Send()
		return nil, err
	}

	actualRecipients := recipients

	// If the recipients list is empty, this signals oneToAll gossip mode,
	// so we use the fanout set as recipients
	switch mode {
	case gossip.ModeOneToAll:
		// recipient list is empty, hence this is a gossip to all (i.e., broadcast)
		gossipMsg.Recipients = nil
		actualRecipients = n.fanoutSet
	case gossip.ModeOneToMany:
		// TODO tentative debug clue
		actualRecipients, err = randomSubSet(recipients, n.staticFanoutNum)
		if err != nil {
			err := errors.Wrap(err, "could not use dynamic fanout, defaulting to one to one")
			log.Debug().Err(err).Send()
			break
		}
	default:
		//dont need to change anything for onetoone
	}

	numberOfRecipients := len(actualRecipients)

	gossipErrChan := make(chan error)
	gossipRepliesChan := make(chan *shared.GossipReply)
	gossipErr := &gossipError{}
	gossipReplies := make([]*shared.GossipReply, numberOfRecipients)

	// TODO for v1.2
	// Note: The order of gossip replies and errors is not defined. Perhaps
	// this is something we need to define for later, as currently there is no clear way
	// to match errors to responses.
	// Loop through recipients
	for _, addr := range actualRecipients {
		go func(addr string) {
			// reply, err := n.placeMessage(ctx, addr, gossipMsg, isSynchronous)
			reply, err := n.placeMessage(context.Background(), addr, gossipMsg, isSynchronous, mode)
			gossipErrChan <- err
			gossipRepliesChan <- reply
		}(addr)
	}

	// Append errors and replies from the respective channels
	for i := 0; i < numberOfRecipients; i++ {
		gossipErr.Append(<-gossipErrChan)
		gossipReplies[i] = <-gossipRepliesChan
	}

	// If encountered errors, then return them to the caller
	if len(*gossipErr) != 0 {
		return gossipReplies, gossipErr
	}

	return gossipReplies, nil
}

// placeMessage is invoked by the node, and encapsulates the process of placing a single message to a recipient's incoming queue
// placeMessage takes a context, address of target node, the message to send, and
// whether to use sync or async mode (true for sync, false for async)
// placeMessage establishes and manages a gRPC client inside
func (n *Node) placeMessage(ctx context.Context, addr string, msg *shared.GossipMessage, isSynchronous bool, mode gossip.Mode) (*shared.GossipReply, error) {
	return n.server.Place(ctx, addr, msg, isSynchronous, mode)
}

// sweeper looks for incoming message to the node's queue, and handles them concurrently
func (n *Node) sweeper() {
	for m := range n.queue {
		go concurrentHandler(n, m)
	}
}

// messageHandler is responsible to handle how to interpret incoming messages.
func (n *Node) messageHandler(ord *order.Order) error {
	log := n.logger.With().
		Str("function", "messageHandler").
		Logger()

	if !order.Valid(ord) {
		err := errors.New("order provided is not valid")
		log.Error().Err(err).Send()
		return err
	}

	invokeResp, err := n.regMngr.Invoke(context.Background(), ord.Msg.GetMessageType(), ord.Msg.GetPayload())
	if err != nil {
		ord.Fill(nil, err)
		err := errors.Wrap(err, "could not invoke")
		log.Debug().Err(err).Send()
		return err
	}

	ord.Fill(invokeResp.Resp, invokeResp.Err)

	return nil
}

// concurrentHandler handles the received messages by gnode in a concurrent manner.
// n: node whose messages are to be handled
// o: order is the order instance managing the message of the node
func concurrentHandler(n *Node, o *order.Order) {
	log := n.logger.With().
		Str("function", "concurrentHandler").
		Logger()
	err := n.messageHandler(o)
	if err != nil {
		err := errors.Wrap(err, "error handling message")
		log.Error().Err(err).Send()
	}
}

// tryStore receives a gossip message and checks if its present inside the node database.
// in case it exists it return true, otherwise it adds the entry to the data base and returns false.
// msg: message to be stored
func (n *Node) tryStore(msg *shared.GossipMessage) (bool, error) {
	hashBytes, err := computeHash(msg)
	if err != nil {
		return false, errors.Wrap(err, "could not compute hash")
	}

	hash := string(hashBytes)

	// if message already received then ignore it.
	if n.hashCache.isReceived(hash) {
		return true, nil
	}

	// add message to storage and set it as recieved
	n.hashCache.receive(hash)
	err = n.msgStore.Put(hash, msg)
	if err != nil {
		return false, errors.Wrap(err, "could not store message")
	}
	return false, nil
}

// pickGossipPartners randomly chooses setSize gossip partners from a given set of peers
func (n *Node) pickGossipPartners() error {
	if len(n.peers) == 0 {
		return errors.New("peers list is empty")
	}

	gossipPartners, err := randomSubSet(n.peers, n.staticFanoutNum)
	if err != nil {
		return errors.Wrap(err, "could not pick gossip partners")
	}

	n.fanoutSet = gossipPartners

	return nil
}
