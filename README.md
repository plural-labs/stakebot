# Autostaker

The autostaker is a server written in Golang for automated multi-chain staking, specifically claiming and delegating on Cosmos SDK chains

## Protocol

Autostaker leverages the `authz` and `feegrant` modules. User accounts must grant the autostaker account access to `MsgDelegate` and `MsgWithdrawDelegatorReward` messages. This service does not include such functionality and must be done prior either via CLI or through a UI.

### Setup

After granting access to perform these messages and cover the fees, Users then register their account as `/register/<account>` i.e. `/register/cosmos1vhpsuaxg51gvvzwyhqejvwfved5ywa3n6vl4ld`. This automatically enables autostaking so long as the chain, in this case `cosmoshub-4` is supported.

### Staking Freuency

The server deploys cron jobs which iteratively claim the accounts rewards, then by calculating transaction fees, delegates the available tokens to the accounts validators, maintaing parity with the percentage delegated. It delegates with a safety margin so that future transactions have sufficient funds. By default, the autostaker operates on a weekly cadence but this can be adjusted

## API

- [POST] `/register/<account>`: Registers an account to the autostakers KV store. Returns an error if the account does not exist or the autostaker doesn't support that chain.
- [POST] `/disable/<account>`: Disables autostaking
- [POST] `/enable/<account>`: Enables autostaking
- [GET] `/status/<account>`: Displays the status of that account
- [POST] `/frequency/daily/<account>`: Sets the update frequency to daily
- [POST] `/frequency/weekly/<account>`: Sets the update frequency to weekly
- [POST] `/frequency/monthly/<account>`: Sets the update frequency to monthly
- [GET] `/address/<chain_id>`: Returns the autostakers address for a specific chain_id. Returns an error if the chain is not supported