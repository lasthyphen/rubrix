package stub

// Hub is a value that stores mocked networks in order for them to send events directly
type Hub struct {
	networks map[string]*Network
	Buffer   *Buffer
}

// NewNetworkHub returns a MockHub value with empty network slice
func NewNetworkHub() *Hub {
	return &Hub{
		networks: make(map[string]*Network),
		Buffer:   NewBuffer(),
	}
}

// GetNetwork returns the Network by the network ID (or node ID)
func (hub *Hub) GetNetwork(networkID string) (*Network, bool) {
	net, ok := hub.networks[networkID]
	return net, ok
}

// Plug stores the reference of the network in the hub object, in order for networks to find
// other network to send events directly
func (hub *Hub) Plug(net *Network) {
	hub.networks[net.GetID()] = net
}