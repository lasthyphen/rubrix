# Consensus 

The consensus stream covers all aspects where final decisions are made about the
network's computation state. 

More specifically, collection covers the following:

**Block Production** 
- Receiving collections from Collector Nodes
- Block formation (generating a candidate block and proposing it to the network)
- BFT consensus to agree on the collections included in a block
- Random Beacon which adds entropy for seeding pseudo-random-number generators to the block
- Block publishing: broadcasting the resulting finalized block to the entire network 

**Block Sealing**  
 - Details TBD (may belong to another stream)


## Terminology

* **Collection** - A set of transactions bundled together by a [Collection Node Cluster](../../../internal/roles/collect)
* **Consensus Node (CN)** - A node that participates in consensus, produces finalized blocks and runs the random beacon.
* **Proto Block** - _Candidate_ blocks (potentially unfinalized) that are produced by the BFT consensus algorithm.
  Proto blocks are full blocks _except_ that they don't contain any entropy (which is subsequently added by the random beacon).
* **Random Beacon** - A _subset_ of consensus nodes that generate entropy through a byzantine-resilient protocol.
  The random beacon adds the entropy (byte-string) to the Proto Blocks.
* (Full) **Block** - A proto block that has been finalized by the BFT consensus protocol and includes entropy generated by the random beacon.
  - has been finalized (committed) by the BFT consensus protocol
  - with added entropy by the Random Beacon       

## Details

### Collection Submission

Collection are submitted to one or more CN(s) via the `SubmitCollection` gRPC method.
  - During normal operation, CNs only receive the collection _hashes_
    and (aggregated) signatures from the collectors that guarantee the collection.  
    The collection's content is _not_ resolved or inspected.  
  - CNs will only consider _guaranteed collections_ (see [Collection](../collect) for details).
Consensus nodes will gossip received collections to other consensus nodes.  


### Block Formation (and Mempool)

Consensus nodes store and track the status of all collections they receive. Each node maintains this sort of mempool for itself.
The mempool only operates on the level of collections and provides the following functionality:
- Provides a list of all pending collections that are _not yet_ included in a finalized block.
- Provides a list of pending collections that are included in non-finalized blocks.
- Orders collections by submission time. 
- For a given fork: provides all pending collections that are not included in this specific fork.
- Collections that are included in a finalized block may be pruned from the mempool.

When a CN generates a proto block, it includes all pending collections that are not in the current fork. 

When a CN sees a proto block from a different CN, it does the following:
* requests and verifies all collections it does not have in its mempool (hashes and signatures only)
* updates the status of all collections in the block 


### BFT Consensus

The BFT consensus algorithm is abstracted behind a generic API.
The only fixed API-level requirement for the consensus algorithm is deterministic finality.    


### Random Beacon

adds entropy for seeding pseudo-random-number generators to the block


### Block Publication

broadcasting the resulting finalized block to the entire network 

## Interaction Graph 
![interaction-flow](./interaction-flow.png?raw=true)