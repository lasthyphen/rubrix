package testnet

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"

	"github.com/dapperlabs/testingdock"

	bootstrapcmd "github.com/dapperlabs/flow-go/cmd/bootstrap/cmd"
	bootstraprun "github.com/dapperlabs/flow-go/cmd/bootstrap/run"
	"github.com/dapperlabs/flow-go/model/bootstrap"
	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/utils/unittest"
)

const (
	// TmpRoot is the default root directory to create temporary data
	// directories for containers. We use /tmp because $TMPDIR is not exposed
	// to docker by default on macOS
	TmpRoot = "/tmp"

	// DefaultBootstrapDir is the default directory for bootstrap files
	DefaultBootstrapDir = "/bootstrap"

	// DefaultFlowDBDir is the default directory for the node database.
	DefaultFlowDBDir = "/flowdb"
	// DefaultExecutionRootDir is the default directory for the execution node
	// state database.
	DefaultExecutionRootDir = "/exedb"

	// ColNodeAPIPort is the name used for the collection node API port.
	ColNodeAPIPort = "col-ingress-port"
	// ExeNodeAPIPort is the name used for the execution node API port.
	ExeNodeAPIPort = "exe-api-port"
	// AccessNodeAPIPort is the name used for the access node API port.
	AccessNodeAPIPort = "access-api-port"
)

func init() {
	testingdock.Verbose = true
	//testingdock.SpawnSequential = true
}

// FlowNetwork represents a test network of Flow nodes running in Docker containers.
type FlowNetwork struct {
	suite       *testingdock.Suite
	config      NetworkConfig
	cli         *dockerclient.Client
	network     *testingdock.Network
	Containers  []*Container
	AccessPorts map[string]string
}

// Identities returns a list of identities, one for each node in the network.
func (net *FlowNetwork) Identities() flow.IdentityList {
	il := make(flow.IdentityList, 0, len(net.Containers))
	for _, c := range net.Containers {
		il = append(il, c.Config.Identity())
	}
	return il
}

// Start starts the network.
func (net *FlowNetwork) Start(ctx context.Context) {
	net.suite.Start(ctx)
}

// Remove stops the network and cleans up all resources. If you need to inspect
// state, first stop the containers, then check state, then clean up resources.
func (net *FlowNetwork) Remove() error {

	err := net.Stop()
	if err != nil {
		return fmt.Errorf("could not stop network: %w", err)
	}

	err = net.Cleanup()
	if err != nil {
		return fmt.Errorf("could not clean up network resources: %w", err)
	}

	return nil
}

// Stop disconnects and stops all containers in the network, then
// removes the network.
func (net *FlowNetwork) Stop() error {

	err := net.suite.Close()
	if err != nil {
		return fmt.Errorf("could not stop containers: %w", err)
	}

	return nil
}

// Cleanup cleans up all temporary files used by the network.
func (net *FlowNetwork) Cleanup() error {

	// remove data directories
	var merr *multierror.Error
	for _, c := range net.Containers {
		err := os.RemoveAll(c.datadir)
		if err != nil {
			merr = multierror.Append(merr, err)
		}
	}

	return merr.ErrorOrNil()
}

// ContainerByID returns the container with the given node ID. If such a
// container exists, returns true. Otherwise returns false.
func (net *FlowNetwork) ContainerByID(id flow.Identifier) (*Container, bool) {
	for _, c := range net.Containers {
		if c.Config.NodeID == id {
			return c, true
		}
	}
	return nil, false
}

// NetworkConfig is the config for the network.
type NetworkConfig struct {
	Nodes     []NodeConfig
	NClusters uint
}

func NewNetworkConfig(nodes []NodeConfig, opts ...func(*NetworkConfig)) NetworkConfig {
	c := NetworkConfig{
		Nodes:     nodes,
		NClusters: 1, // default to 1 cluster
	}

	for _, apply := range opts {
		apply(&c)
	}

	return c
}

func WithClusters(n uint) func(*NetworkConfig) {
	return func(conf *NetworkConfig) {
		conf.NClusters = n
	}
}

func (n *NetworkConfig) Len() int {
	return len(n.Nodes)
}

func (n *NetworkConfig) Less(i, j int) bool {
	return n.Nodes[i].Role < n.Nodes[j].Role
}

func (n *NetworkConfig) Swap(i, j int) {
	n.Nodes[i], n.Nodes[j] = n.Nodes[j], n.Nodes[i]
}

// NodeConfig defines the input config for a particular node, specified prior
// to network creation.
type NodeConfig struct {
	Role       flow.Role
	Stake      uint64
	Identifier flow.Identifier
	LogLevel   zerolog.Level
}

func NewNodeConfig(role flow.Role, opts ...func(*NodeConfig)) NodeConfig {
	c := NodeConfig{
		Role:       role,
		Stake:      1000,                         // default stake
		Identifier: unittest.IdentifierFixture(), // default random ID
		LogLevel:   zerolog.DebugLevel,           // log at debug by default
	}

	for _, apply := range opts {
		apply(&c)
	}

	return c
}

// NewNodeConfigSet creates a set of node configs with the given role. The nodes
// are given sequential IDs with a common prefix to make reading logs easier.
func NewNodeConfigSet(n int, role flow.Role, opts ...func(*NodeConfig)) []NodeConfig {

	// each node in the set has a common 4-digit prefix, separated from their
	// index with a `0` character
	idPrefix := rand.Intn(10000) * 100

	confs := make([]NodeConfig, n)
	for i := 0; i < n; i++ {
		confs[i] = NewNodeConfig(role, append(opts, WithIDInt(uint(idPrefix+i+1)))...)
	}

	return confs
}

func WithID(id flow.Identifier) func(config *NodeConfig) {
	return func(config *NodeConfig) {
		config.Identifier = id
	}
}

// WithIDInt sets the node ID so the hex representation matches the input.
// Useful for having consistent and easily readable IDs in test logs.
func WithIDInt(id uint) func(config *NodeConfig) {

	idStr := strconv.Itoa(int(id))
	// left pad ID with zeros
	pad := strings.Repeat("0", 64-len(idStr))
	hex := pad + idStr

	// convert hex to ID
	flowID, err := flow.HexStringToIdentifier(hex)
	if err != nil {
		panic(err)
	}

	return WithID(flowID)
}

func WithLogLevel(level zerolog.Level) func(config *NodeConfig) {
	return func(config *NodeConfig) {
		config.LogLevel = level
	}
}

func PrepareFlowNetwork(t *testing.T, name string, networkConf NetworkConfig) (*FlowNetwork, error) {

	// number of nodes
	nNodes := len(networkConf.Nodes)

	if nNodes == 0 {
		return nil, fmt.Errorf("must specify at least one node")
	}

	// Sort so that access nodes start up last
	sort.Sort(&networkConf)

	// set up docker client
	dockerClient, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	require.Nil(t, err)

	suite, _ := testingdock.GetOrCreateSuite(t, name, testingdock.SuiteOpts{
		Client: dockerClient,
	})
	network := suite.Network(testingdock.NetworkOpts{
		Name: name,
	})

	// generate staking and networking keys for each configured node
	confs := setupKeys(t, networkConf)

	// run DKG for all consensus nodes
	dkg := runDKG(t, confs)

	// generate genesis block
	seal := bootstraprun.GenerateRootSeal(flow.GenesisStateCommitment)
	genesis := bootstraprun.GenerateRootBlock(toIdentityList(confs), seal)

	// generate QC
	nodeInfos := bootstrap.FilterByRole(toNodeInfoList(confs), flow.RoleConsensus)
	signerData := bootstrapcmd.GenerateQCParticipantData(nodeInfos, nodeInfos, dkg)
	qc, err := bootstraprun.GenerateGenesisQC(signerData, &genesis)
	require.Nil(t, err)

	// create a temporary directory to store all bootstrapping files, these
	// will be shared between all nodes
	bootstrapDir, err := ioutil.TempDir(TmpRoot, "flow-integration-bootstrap")
	require.Nil(t, err)

	// write common genesis bootstrap files
	err = writeJSON(filepath.Join(bootstrapDir, bootstrap.FilenameGenesisBlock), genesis)
	require.Nil(t, err)
	err = writeJSON(filepath.Join(bootstrapDir, bootstrap.FilenameGenesisQC), qc)
	require.Nil(t, err)
	err = writeJSON(filepath.Join(bootstrapDir, bootstrap.FilenameDKGDataPub), dkg.Public())
	require.Nil(t, err)

	// write private key files for each DKG participant
	for _, part := range dkg.Participants {
		filename := fmt.Sprintf(bootstrap.FilenameRandomBeaconPriv, part.NodeID)
		err = writeJSON(filepath.Join(bootstrapDir, filename), part.Private())
		require.Nil(t, err)
	}

	// write private key files for each node
	for _, nodeConfig := range confs {
		path := filepath.Join(bootstrapDir, fmt.Sprintf(bootstrap.FilenameNodeInfoPriv, nodeConfig.NodeID))

		// retrieve private representation of the node
		private, err := nodeConfig.NodeInfo.Private()
		require.Nil(t, err)

		err = writeJSON(path, private)
		require.Nil(t, err)
	}

	flowNetwork := &FlowNetwork{
		cli:         dockerClient,
		config:      networkConf,
		suite:       suite,
		network:     network,
		Containers:  make([]*Container, 0, nNodes),
		AccessPorts: make(map[string]string),
	}

	// add each node to the network
	for _, nodeConf := range confs {
		err = flowNetwork.AddNode(t, bootstrapDir, nodeConf)
		require.Nil(t, err)
	}

	return flowNetwork, nil
}

// AddNode creates a node container with the given config and adds it to the
// network.
func (net *FlowNetwork) AddNode(t *testing.T, bootstrapDir string, nodeConf ContainerConfig) error {

	opts := &testingdock.ContainerOpts{
		ForcePull: false,
		Name:      nodeConf.ContainerName,
		Config: &container.Config{
			Image: nodeConf.ImageName(),
			User:  currentUser(),
			Cmd: []string{
				fmt.Sprintf("--nodeid=%s", nodeConf.NodeID.String()),
				fmt.Sprintf("--bootstrapdir=%s", DefaultBootstrapDir),
				fmt.Sprintf("--datadir=%s", DefaultFlowDBDir),
				fmt.Sprintf("--loglevel=%s", nodeConf.LogLevel),
				fmt.Sprintf("--nclusters=%d", net.config.NClusters),
			},
		},
		HostConfig: &container.HostConfig{},
	}

	// get a temporary directory in the host. On macOS the default tmp
	// directory is NOT accessible to Docker by default, so we use /tmp
	// instead.
	tmpdir, err := ioutil.TempDir(TmpRoot, "flow-integration-node")
	if err != nil {
		return fmt.Errorf("could not get tmp dir: %w", err)
	}

	nodeContainer := &Container{
		Config:  nodeConf,
		Ports:   make(map[string]string),
		datadir: tmpdir,
		net:     net,
		opts:    opts,
	}

	// create a directory for the node database
	flowDBDir := filepath.Join(tmpdir, DefaultFlowDBDir)
	err = os.Mkdir(flowDBDir, 0700)
	require.Nil(t, err)

	// Bind the host directory to the container's database directory
	// Bind the common bootstrap directory to the container
	// NOTE: I did this using the approach from:
	// https://github.com/fsouza/go-dockerclient/issues/132#issuecomment-50694902
	opts.HostConfig.Binds = append(
		opts.HostConfig.Binds,
		fmt.Sprintf("%s:%s:rw", flowDBDir, DefaultFlowDBDir),
		fmt.Sprintf("%s:%s:ro", bootstrapDir, DefaultBootstrapDir),
	)

	switch nodeConf.Role {
	case flow.RoleCollection:

		hostPort := testingdock.RandomPort(t)
		containerPort := "9000/tcp"

		nodeContainer.bindPort(hostPort, containerPort)

		nodeContainer.addFlag("ingress-addr", fmt.Sprintf("%s:9000", nodeContainer.Name()))
		nodeContainer.opts.HealthCheck = testingdock.HealthCheckCustom(healthcheckAccessGRPC(hostPort))
		nodeContainer.Ports[ColNodeAPIPort] = hostPort
		net.AccessPorts[ColNodeAPIPort] = hostPort
	case flow.RoleExecution:

		hostPort := testingdock.RandomPort(t)
		containerPort := "9000/tcp"

		nodeContainer.bindPort(hostPort, containerPort)

		nodeContainer.addFlag("rpc-addr", fmt.Sprintf("%s:9000", nodeContainer.Name()))
		nodeContainer.opts.HealthCheck = testingdock.HealthCheckCustom(healthcheckExecutionGRPC(hostPort))
		nodeContainer.Ports[ExeNodeAPIPort] = hostPort
		net.AccessPorts[ExeNodeAPIPort] = hostPort

		// create directories for execution state trie and values in the tmp
		// host directory.
		tmpLedgerDir, err := ioutil.TempDir(tmpdir, "flow-integration-trie")
		require.Nil(t, err)

		opts.HostConfig.Binds = append(
			opts.HostConfig.Binds,
			fmt.Sprintf("%s:%s:rw", tmpLedgerDir, DefaultExecutionRootDir),
		)

		nodeContainer.addFlag("triedir", DefaultExecutionRootDir)
	}

	suiteContainer := net.suite.Container(*opts)
	net.network.After(suiteContainer)
	nodeContainer.Container = suiteContainer
	net.Containers = append(net.Containers, nodeContainer)

	return nil
}

// setupKeys generates private staking and networking keys for each configured
// node. It also assigns each node a unique container name and network address.
func setupKeys(t *testing.T, networkConf NetworkConfig) []ContainerConfig {

	nNodes := len(networkConf.Nodes)

	// keep track of how many roles we have assigned so we can number containers
	// correctly (consensus_1, consensus_2, etc.)
	roleCounter := make(map[flow.Role]int)

	// get networking keys for all nodes
	networkKeys, err := unittest.NetworkingKeys(nNodes)
	require.Nil(t, err)

	// get staking keys for all nodes
	stakingKeys, err := unittest.StakingKeys(nNodes)
	require.Nil(t, err)

	// create node container configs and corresponding public identities
	confs := make([]ContainerConfig, 0, nNodes)
	for i, conf := range networkConf.Nodes {

		// define the node's name <role>_<n> and address <name>:<port>
		name := fmt.Sprintf("%s_%d", conf.Role.String(), roleCounter[conf.Role]+1)

		addr := fmt.Sprintf("%s:%d", name, 2137)
		roleCounter[conf.Role]++

		info := bootstrap.NewPrivateNodeInfo(
			conf.Identifier,
			conf.Role,
			addr,
			conf.Stake,
			networkKeys[i],
			stakingKeys[i],
		)

		containerConf := ContainerConfig{
			NodeInfo:      info,
			ContainerName: name,
			LogLevel:      conf.LogLevel.String(),
		}

		confs = append(confs, containerConf)
	}

	return confs
}

func runDKG(t *testing.T, confs []ContainerConfig) bootstrap.DKGData {

	// filter by consensus nodes
	consensusNodes := bootstrap.FilterByRole(toNodeInfoList(confs), flow.RoleConsensus)
	nConsensusNodes := len(consensusNodes)

	// run the core dkg algorithm
	dkgSeeds, err := getSeeds(nConsensusNodes)
	require.Nil(t, err)
	dkg, err := bootstraprun.RunDKG(nConsensusNodes, dkgSeeds)
	require.Nil(t, err)

	// sanity check
	assert.Equal(t, nConsensusNodes, len(dkg.Participants))

	// set the node IDs in the dkg data
	for i := range dkg.Participants {
		nodeID := consensusNodes[i].NodeID
		dkg.Participants[i].NodeID = nodeID
	}

	return dkg
}
