package gotezos

import (
	"github.com/DefinitelyNotAGoat/go-tezos/v2/account"
	"github.com/DefinitelyNotAGoat/go-tezos/v2/block"
	tzc "github.com/DefinitelyNotAGoat/go-tezos/v2/client"
	"github.com/DefinitelyNotAGoat/go-tezos/v2/contracts"
	"github.com/DefinitelyNotAGoat/go-tezos/v2/cycle"
	"github.com/DefinitelyNotAGoat/go-tezos/v2/delegate"
	"github.com/DefinitelyNotAGoat/go-tezos/v2/network"
	"github.com/DefinitelyNotAGoat/go-tezos/v2/node"
	"github.com/DefinitelyNotAGoat/go-tezos/v2/operations"
	"github.com/DefinitelyNotAGoat/go-tezos/v2/snapshot"
	"github.com/pkg/errors"
)

// GoTezos is the driver of the library, it inludes the several RPC services
// like Block, SnapSHot, Cycle, Account, Delegate, Operations, Contract, and Network
type GoTezos struct {
	Client    tzc.TezosClient
	Constants network.Constants
	Block     block.TezosBlockService
	Snapshot  snapshot.TezosSnapshotService
	Cycle     cycle.TezosCycleService
	Account   account.TezosAccountService
	Delegate  delegate.TezosDelegateService
	Network   network.TezosNetworkService
	Operation operations.TezosOperationsService
	Contract  contracts.TezosContractsService
	Node      node.TezosNodeService
}

// NewGoTezos is a constructor that returns a GoTezos object
func NewGoTezos(URL string) (*GoTezos, error) {
	gt := GoTezos{}

	gt.Client = tzc.NewClient(URL)
	gt.Network = network.NewNetworkService(gt.Client)
	var err error
	gt.Constants, err = gt.Network.GetConstants()
	if err != nil {
		return &gt, errors.Wrap(err, "could not get network constants")
	}
	gt.Block = block.NewBlockService(gt.Client)
	gt.Cycle = cycle.NewCycleService(gt.Block)
	gt.Snapshot = snapshot.NewSnapshotService(
		gt.Cycle,
		gt.Client,
		gt.Block,
		gt.Constants,
	)
	gt.Account = account.NewAccountService(
		gt.Client,
		gt.Block,
		gt.Snapshot,
	)
	gt.Delegate = delegate.NewDelegateService(
		gt.Client,
		gt.Block,
		gt.Snapshot,
		gt.Account,
		gt.Constants,
	)
	gt.Operation = operations.NewOperationService(gt.Block, gt.Client)
	gt.Contract = contracts.NewContractService(gt.Client)
	gt.Node = node.NewNodeService(gt.Client)

	return &gt, nil
}
