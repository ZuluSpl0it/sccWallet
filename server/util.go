package server

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/wallet"
	"gitlab.com/scpcorp/ScPrime/types"
)

// For an unconfirmed Transaction, the TransactionTimestamp field is set to the
// maximum value of a uint64.
const unconfirmedTransactionTimestamp = ^uint64(0)

// ComputeSummarizedTransactions creates a set of SummarizedTransactions
// from a set of ProcessedTransactions.
func ComputeSummarizedTransactions(pts []modules.ProcessedTransaction, blockHeight types.BlockHeight) ([]modules.SummarizedTransaction, error) {
	sts := []modules.SummarizedTransaction{}
	vts, err := wallet.ComputeValuedTransactions(pts, blockHeight)
	if err != nil {
		return nil, err
	}
	for _, txn := range vts {
		// Determine the number of outgoing coins and funds.
		var outgoingFunds types.Currency
		for _, input := range txn.Inputs {
			if input.FundType == types.SpecifierSiafundInput && input.WalletAddress {
				outgoingFunds = outgoingFunds.Add(input.Value)
			}
		}
		// Determine the number of incoming funds.
		var incomingFunds types.Currency
		for _, output := range txn.Outputs {
			if output.FundType == types.SpecifierSiafundOutput && output.WalletAddress {
				incomingFunds = incomingFunds.Add(output.Value)
			}
		}
		// Convert the scp to a float.
		incomingCoinsFloat, _ := new(big.Rat).SetFrac(txn.ConfirmedIncomingValue.Big(), types.ScPrimecoinPrecision.Big()).Float64()
		outgoingCoinsFloat, _ := new(big.Rat).SetFrac(txn.ConfirmedOutgoingValue.Big(), types.ScPrimecoinPrecision.Big()).Float64()
		// Summarize transaction
		st := modules.SummarizedTransaction{}
		st.TxnId = strings.ToUpper(fmt.Sprintf("%v", txn.TransactionID))
		st.Type = strings.ToUpper(strings.Replace(fmt.Sprintf("%v", txn.TxType), "_", " ", -1))
		if uint64(txn.ConfirmationTimestamp) != unconfirmedTransactionTimestamp {
			st.Time = time.Unix(int64(txn.ConfirmationTimestamp), 0).Format("2006-01-02 15:04")
			st.Confirmed = "Yes"
		} else {
			st.Confirmed = "No"
		}
		st.Scp = fmt.Sprintf("%15.2f SCP", incomingCoinsFloat-outgoingCoinsFloat)
		// For funds, need to avoid having a negative types.Currency.
		if incomingFunds.Cmp(outgoingFunds) > 0 {
			st.Spf = fmt.Sprintf("%14v SPF %v\n", incomingFunds.Sub(outgoingFunds), txn.TxType)
		} else if incomingFunds.Cmp(outgoingFunds) < 0 {
			st.Spf = fmt.Sprintf("-%14v SPF %v\n", outgoingFunds.Sub(incomingFunds), txn.TxType)
		}
		sts = append(sts, st)
	}
	return sts, nil
}
