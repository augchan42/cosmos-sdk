package slashing

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// Iterate over all the validators which *should* have signed this block
	// store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)
	for _, voteInfo := range req.LastCommitInfo.GetVotes() {
		k.HandleValidatorSignature(ctx, voteInfo.Validator.Address, voteInfo.Validator.Power, voteInfo.SignedLastBlock)
	}
}

// Called every block, update validator set
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	if ctx.BlockHeight()%10 == 0 { // TODO should update hardcoded 10 to stakingParams.EpochInterval (epoch_interval)
		// execute all epoch actions
		iterator := k.GetEpochActionsIterator()
	
		for ; iterator.Valid(); iterator.Next() {
			var msg sdk.Msg
			bz := iterator.Value()
			k.cdc.MustUnmarshalBinaryBare(bz, &msg)

			switch msg := msg.(type) {
			case *types.MsgUnjail:
				res, err := k.EpochUnjail(sdk.WrapSDKContext(ctx), msg)

			default:
			}
		}
		// dequeue all epoch actions after run
		k.DequeueEpochActions()
	}
	return []abci.ValidatorUpdate{}
}
