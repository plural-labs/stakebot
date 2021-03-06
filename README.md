# Stakebot

Stakebot is a server written in Golang for automated multi-chain staking, specifically claiming and delegating on Cosmos SDK chains. For the client implementations look at the [autostaker](https://github.com/plural-labs/autostaker).

## Protocol

Stakebot leverages the `authz` and `feegrant` modules. User accounts must grant the stakebot account access to `MsgDelegate` and `MsgWithdrawDelegatorReward` messages. This service does not include such functionality and must be done prior either via CLI or through a UI.

### Staking Frequency

The server deploys cron jobs which iteratively claim the accounts rewards, then by calculating transaction fees, delegates the available balance of native tokens to the accounts validators, maintaing parity with the percentage delegated. It delegates with a safety margin known as `tolerance` so that future transactions have sufficient funds. The stakebot has defaults per chain for `frequency` and `tolerance` but the frequency this can be adjusted to one of following:

1. `hourly`
2. `quarterday`
3. `daily`
4. `weekly`
5. `monthly`

## Setup

1. Clone the repo `git clone https://github.com/plural-labs/stakebot`.
2. Install the `stakebot`: `go install` from the root directory.
3. Run `stakebot init` to create a set of keys and and the default config.
4. Move into the directory `cd ~/.stakebot` and edit the config `vim config.toml`, adding chain details for the chains you want to support.
5. Run `stakebot serve` to begin the server. You will see some logs on start up.

A few extra utility commands:

- `stakebot address` returns the address of the server
- `stakebot find <address>` can be used to get info on a particular address stakebot is serving.

## Usage

To set up autostaking via the REST server:

1. Get the address of the stakebot by calling `/v1/address?id=<chain_id>`.
2. Manually grant the address the authority to call the two aforementioned msg types as well as a feegrant.
3. After granting access to perform the messages and cover the fees, register your account as `/v1/register?address=<account>&frequeny=<frequency>&tolerance=<tolerance>` i.e. `/v1/register?address=cosmos1vhpsuaxg51gvvzwyhqejvwfved5ywa3n6vl4ld`. This automatically enables autostaking so long as the chain, in this case `cosmoshub-4` is supported. If you don't add a `frequency` or `tolerance`, reasonable defaults will be chosen from the server settings.
4. If you want to manually trigger a restake you can also run: `/v1/restake?address=<address>`.

Alternatively, checkout the [autostaker](https://github.com/plural-labs/autostaker) CLI and frontend

## API Reference

- `/v1/register?address=<account>`: Registers an account to the stakebot's KV store. Returns an error if the account does not exist or the stakebot doesn't support that chain.
- `/v1/restake?address=<account>`: Manu
- `/v1/status?address=<account>`: Displays the status of that account
- `/v1/chains`: Returns all chains that the stakebot server supports
- `/v1/chain?id=<chain_id>`: Returns information on the specified chain if the stakebot server supports it.
- `/address/<chain_id>`: Returns the stakebot's address for a specific chain_id. Returns an error if the chain is not supported.
