#!/bin/bash

KEYS=("validator" "test")
CHAINID="devnet"
MONIKER="Side Labs"
BINARY="$HOME/go/bin/sided"
DENOM_STR="uside,sat,uusdc,uusdt"
INITIAL_ACCOUNT_STR=""
set -f
IFS=,
DENOMS=($DENOM_STR)
INITIAL_ACCOUNTS=($INITIAL_ACCOUNT_STR)

IFS=";"

INITIAL_SUPPLY="500000000000000"
BLOCK_GAS=10000000
MAX_GAS=10000000000

# btcbridge params
BTC_VAULT=() # ("<address>" "<pk>" "<asset type>")
RUNES_VAULT=()
TRUSTED_BTC_RELAYER=""
TRUSTED_NON_BTC_RELAYER=""
TRUSTED_ORACLE=""
PROTOCOL_FEE_COLLECTOR=""

# gov params
GOV_VOTING_PERIOD="600s"
GOV_EXPEDITED_VOTING_PERIOD="300s"

# Remember to change to other types of keyring like 'file' in-case exposing to outside world,
# otherwise your balance will be wiped quickly
# The keyring test does not require private key to steal tokens from you
KEYRING="test"
#KEYALGO="secp256k1"
KEYALGO="segwit"
LOGLEVEL="info"
# Set dedicated home directory for the $BINARY instance
HOMEDIR="$HOME/.side"

# Path variables
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

# Reinstall daemon
make install

# User prompt if an existing local node configuration is found.
if [ -d "$HOMEDIR" ]; then
	printf "\nAn existing folder at '%s' was found. You can choose to delete this folder and start a new local node with new keys from genesis. When declined, the existing local node is started. \n" "$HOMEDIR"
	echo "Overwrite the existing configuration and start a new local node? [y/n]"
	read -r overwrite
else
	overwrite="Y"
fi


# Setup local node if overwrite is set to Yes, otherwise skip setup
if [[ $overwrite == "y" || $overwrite == "Y" ]]; then
	# Remove the previous folder
	rm -rf "$HOMEDIR"

	# Set client config
	$BINARY config keyring-backend $KEYRING --home "$HOMEDIR"
	$BINARY config chain-id $CHAINID --home "$HOMEDIR"

	# If keys exist they should be deleted
	for KEY in "${KEYS[@]}"; do
		$BINARY keys add "$KEY" --keyring-backend $KEYRING --algo $KEYALGO --home "$HOMEDIR"
	done
	# for KEY in "${KEYS[@]}"; do
    # # Add the --recover flag to initiate recovery mode
    # 	$BINARY keys add "$KEY" --keyring-backend $KEYRING --algo $KEYALGO --recover --home "$HOMEDIR"
	# done

	echo ""
	echo "☝️ Copy the above mnemonic phrases and import them to relayer! Press [Enter] to continue..."
	read -r continue

	# Set moniker and chain-id for Cascadia (Moniker can be anything, chain-id must be an integer)
	$BINARY init $MONIKER -o --chain-id $CHAINID --home "$HOMEDIR"

	jq --arg denom "${DENOMS[0]}" '.app_state["staking"]["params"]["bond_denom"]=$denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq --arg denom "${DENOMS[0]}" '.app_state["mint"]["params"]["mint_denom"]=$denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq --arg denom "${DENOMS[0]}" '.app_state["crisis"]["constant_fee"]["denom"]=$denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq --arg denom "${DENOMS[0]}" '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]=$denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq --arg denom "${DENOMS[0]}" '.app_state["gov"]["params"]["min_deposit"][0]["denom"]=$denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	# Set gas limit in genesis
	jq --arg max_gas "$MAX_GAS" '.consensus_params["block"]["max_gas"]=$max_gas' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

    # set gov voting period
	if [ -n "$GOV_VOTING_PERIOD" ]; then
        jq --arg voting_period "${GOV_VOTING_PERIOD}" '.app_state["gov"]["params"]["voting_period"]=$voting_period' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	fi

	# set gov expedited voting period
	if [ -n "$GOV_EXPEDITED_VOTING_PERIOD" ]; then
        jq --arg expedited_voting_period "${GOV_EXPEDITED_VOTING_PERIOD}" '.app_state["gov"]["params"]["expedited_voting_period"]=$expedited_voting_period' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	fi

	# set vaults if provided
    if [[ "${#BTC_VAULT[@]}" -eq 3 && "${#RUNES_VAULT[@]}" -eq 3 ]]; then
	    jq --arg btc_vault "${BTC_VAULT[0]}" '.app_state["btcbridge"]["params"]["vaults"][0]["address"]=$btc_vault' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
		jq --arg btc_vault_pk "${BTC_VAULT[1]}" '.app_state["btcbridge"]["params"]["vaults"][0]["pub_key"]=$btc_vault_pk' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
		jq --arg btc_vault_asset_type "${BTC_VAULT[2]}" '.app_state["btcbridge"]["params"]["vaults"][0]["asset_type"]=$btc_vault_asset_type' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
		jq --arg runes_vault "${RUNES_VAULT[0]}" '.app_state["btcbridge"]["params"]["vaults"][1]["address"]=$runes_vault' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
		jq --arg runes_vault_pk "${RUNES_VAULT[1]}" '.app_state["btcbridge"]["params"]["vaults"][1]["pub_key"]=$runes_vault_pk' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
		jq --arg runes_vault_asset_type "${RUNES_VAULT[2]}" '.app_state["btcbridge"]["params"]["vaults"][1]["asset_type"]=$runes_vault_asset_type' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
    fi

    # set trusted btc relayer
	if [ -n "$TRUSTED_BTC_RELAYER" ]; then
	    jq --arg relayer "$TRUSTED_BTC_RELAYER" '.app_state["btcbridge"]["params"]["trusted_btc_relayers"][0]=$relayer' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
    fi

	# set trusted non btc relayer
	if [ -n "$TRUSTED_NON_BTC_RELAYER" ]; then
	    jq --arg relayer "$TRUSTED_NON_BTC_RELAYER" '.app_state["btcbridge"]["params"]["trusted_non_btc_relayers"][0]=$relayer' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
    fi

	# set trusted oracle
	if [ -n "$TRUSTED_ORACLE" ]; then
	    jq --arg oracle "$TRUSTED_ORACLE" '.app_state["btcbridge"]["params"]["trusted_oracles"][0]=$oracle' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
    fi

	# set protocol fee collector
	if [ -n "$PROTOCOL_FEE_COLLECTOR" ]; then
	    jq --arg fee_collector "$PROTOCOL_FEE_COLLECTOR" '.app_state["btcbridge"]["params"]["protocol_fees"]["collector"]=$fee_collector' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
    fi

	# set custom pruning settings
	sed -i.bak 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
	sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
	sed -i.bak 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"

	sed -i.bak 's/127.0.0.1:26657/0.0.0.0:26657/g' "$CONFIG"
	sed -i.bak 's/cors_allowed_origins\s*=\s*\[\]/cors_allowed_origins = ["*",]/g' "$CONFIG"
	sed -i.bak 's/swagger = false/swagger = true/g' $APP_TOML

	# Allocate genesis accounts (cosmos formatted addresses)
	for KEY in "${KEYS[@]}"; do
	    BALANCES=""
	    for key in "${!DENOMS[@]}"; do
	        BALANCES+=",${INITIAL_SUPPLY}${DENOMS[$key]}"
	    done
	    echo ${BALANCES:1}
	    $BINARY genesis add-genesis-account "$KEY" ${BALANCES:1} --keyring-backend $KEYRING --home "$HOMEDIR"
	done

	echo "Genesis accounts allocated for local accounts"

	# Allocate genesis accounts (cosmos formatted addresses)
	for ADDR in "${INITIAL_ACCOUNTS[@]}"; do
	    BALANCES=""
	    for key in "${!DENOMS[@]}"; do
	        BALANCES+=",${INITIAL_SUPPLY}${DENOMS[$key]}"
	    done
	    echo ${BALANCES:1}
	    $BINARY genesis add-genesis-account "$ADDR" ${BALANCES:1} --home "$HOMEDIR"
	done
	echo "Genesis accounts allocated for initial accounts"



	# Sign genesis transaction
	# echo $INITIAL_SUPPLY${DENOMS[0]}
	$BINARY genesis gentx "${KEYS[0]}" $INITIAL_SUPPLY${DENOMS[0]} --keyring-backend $KEYRING --chain-id $CHAINID --identity "666AC57CC678BEC4" --website="https://side.one" --home "$HOMEDIR"
	echo "Genesis transaction signed"

	## In case you want to create multiple validators at genesis
	## 1. Back to `$BINARY keys add` step, init more keys
	## 2. Back to `$BINARY add-genesis-account` step, add balance for those
	## 3. Clone this ~/.$BINARY home directory into some others, let's say `~/.clonedCascadiad`
	## 4. Run `gentx` in each of those folders
	## 5. Copy the `gentx-*` folders under `~/.clonedCascadiad/config/gentx/` folders into the original `~/.$BINARY/config/gentx`

	# Collect genesis tx
	$BINARY genesis collect-gentxs --home "$HOMEDIR"
	echo "Genesis transactions collected"

	# Run this to ensure everything worked and that the genesis file is setup correctly
	$BINARY genesis validate --home "$HOMEDIR"
	echo "Genesis file validated"

	if [[ $1 == "pending" ]]; then
		echo "pending mode is on, please wait for the first block committed."
	fi
fi


# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
$BINARY start --log_level info --minimum-gas-prices=0.0001${DENOMS[0]} --home "$HOMEDIR"
