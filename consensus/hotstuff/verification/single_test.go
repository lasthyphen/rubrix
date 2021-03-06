package verification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/consensus/hotstuff/helper"
	"github.com/onflow/flow-go/consensus/hotstuff/model"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/utils/unittest"
)

func TestSingleVote(t *testing.T) {

	identities := unittest.IdentityListFixture(4, unittest.WithRole(flow.RoleConsensus))
	committeeState, stakingKeys, _ := MakeHotstuffCommitteeState(t, identities, false, epochCounter)
	signers := MakeSigners(t, committeeState, identities.NodeIDs(), stakingKeys, nil)

	// create proposal
	block := helper.MakeBlock(t, helper.WithBlockProposer(identities[2].NodeID))
	vote, err := signers[0].CreateVote(block)
	require.NoError(t, err)
	voter := identities[0]

	// vote should be valid
	valid, err := signers[0].VerifyVote(voter, vote.SigData, block)
	require.NoError(t, err)
	assert.True(t, valid, "original vote should be valid")

	// vote on different block should be invalid
	block.BlockID[0]++
	valid, err = signers[0].VerifyVote(voter, vote.SigData, block)
	require.NoError(t, err)
	assert.False(t, valid, "vote with changed block ID should be invalid")
	block.BlockID[0]--

	// vote with changed view should be invalid
	block.View++
	valid, err = signers[0].VerifyVote(voter, vote.SigData, block)
	require.NoError(t, err)
	assert.False(t, valid, "vote with changed view should be invalid")
	block.View--

	// vote by different signer should be invalid
	voter = identities[1]
	valid, err = signers[0].VerifyVote(voter, vote.SigData, block)
	require.NoError(t, err)
	assert.False(t, valid, "vote with changed identity should be invalid")
	voter = identities[0]

	// vote with changed signature should be invalid
	vote.SigData[0]++
	valid, err = signers[0].VerifyVote(voter, vote.SigData, block)
	require.NoError(t, err)
	assert.False(t, valid, "vote with changed signature should be invalid")
	vote.SigData[0]--
}

func TestSingleProposalIsVote(t *testing.T) {

	// NOTE: I don't think this is true for every signature scheme

	identities := unittest.IdentityListFixture(4, unittest.WithRole(flow.RoleConsensus))
	committeeState, stakingKeys, _ := MakeHotstuffCommitteeState(t, identities, false, epochCounter)
	signers := MakeSigners(t, committeeState, identities.NodeIDs(), stakingKeys, nil)

	// create proposal
	block := helper.MakeBlock(t, helper.WithBlockProposer(identities[0].NodeID))
	proposal, err := signers[0].CreateProposal(block)
	require.NoError(t, err)
	vote, err := signers[0].CreateVote(block)
	require.NoError(t, err)

	assert.Equal(t, proposal.SigData, vote.SigData)
}

func TestSingleQC(t *testing.T) {

	identities := unittest.IdentityListFixture(4, unittest.WithRole(flow.RoleConsensus))
	minShares := (len(identities)-1)/2 + 1
	committeeState, stakingKeys, _ := MakeHotstuffCommitteeState(t, identities, false, epochCounter)
	signers := MakeSigners(t, committeeState, identities.NodeIDs(), stakingKeys, nil)

	// create proposal
	block := helper.MakeBlock(t, helper.WithBlockProposer(identities[0].NodeID))
	var votes []*model.Vote
	for _, signer := range signers {
		vote, err := signer.CreateVote(block)
		require.NoError(t, err)
		votes = append(votes, vote)
	}

	// should be able to create QC from votes and verify
	qc, err := signers[0].CreateQC(votes[:minShares])
	require.NoError(t, err, "should be able to create QC from valid votes")

	// creation from different views should fail
	votes[0].View++
	_, err = signers[0].CreateQC(votes[:minShares])
	assert.Error(t, err, "creating QC with mismatching view should fail")
	votes[0].View--

	// creation from different block IDs should fail
	votes[0].BlockID[0]++
	_, err = signers[0].CreateQC(votes[:minShares])
	assert.Error(t, err, "creating QC with mismatching block ID should fail")
	votes[0].BlockID[0]--

	// should be able to verify valid QC
	valid, err := signers[0].VerifyQC(identities[:minShares], qc.SigData, block)
	require.NoError(t, err)
	assert.True(t, valid, "original QC should be valid")

	// Verify QC with a totally different set of signers (to test the staking key aggregation)
	qc_bis, err := signers[0].CreateQC(votes[len(votes)-minShares:])
	require.NoError(t, err, "should be able to create QC from valid votes")
	valid, err = signers[0].VerifyQC(identities[len(identities)-minShares:], qc_bis.SigData, block)
	require.NoError(t, err)
	assert.True(t, valid, "QC should be valid")

	// Verify QC with a slightly different set of signers (to test the staking key aggregation)
	for i := 0; i <= len(votes)-minShares; i++ {
		qc_bis, err = signers[0].CreateQC(votes[i : i+minShares])
		require.NoError(t, err, "should be able to create QC from valid votes")
		valid, err = signers[0].VerifyQC(identities[i:i+minShares], qc_bis.SigData, block)
		require.NoError(t, err)
		assert.True(t, valid, "QC should be valid")
	}

	// verification with with not enough voters is invalid
	valid, err = signers[0].VerifyQC(identities[:minShares-1], qc.SigData, block)
	require.NoError(t, err)
	assert.False(t, valid, "verification with missing voter ID should not work")

	// verification with changed signature should fail
	// TODO: change error handling so split failure and invalid signature are
	// treated the same
	qc.SigData[4]++
	valid, err = signers[0].VerifyQC(identities[:minShares], qc.SigData, block)
	require.NoError(t, err)
	assert.False(t, valid, "QC with changed signature data should be invalid")
	qc.SigData[4]--

	// verification with changed block ID should fail
	block.BlockID[0]++
	valid, err = signers[0].VerifyQC(identities[:minShares], qc.SigData, block)
	require.NoError(t, err)
	assert.False(t, valid, "QC with changed block ID data should be invalid")
	block.BlockID[0]--

	// verification with changed view should fail
	block.View++
	valid, err = signers[0].VerifyQC(identities[:minShares], qc.SigData, block)
	require.NoError(t, err)
	assert.False(t, valid, "QC with changed block view data should be invalid")
	block.View--
}
