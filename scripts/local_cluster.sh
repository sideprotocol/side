#!/bin/bash

N=4
START=2 # start all node
# START=3 # skip validator 2
if [[ -n "$1" ]]; then
    START=$((START + $1))
fi

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

INITIAL_SUPPLY="100000000000000"
BLOCK_GAS=10000000
MAX_GAS=10000000000

# btcbridge params
BTC_VAULT=() # ("<address>" "<pk>" "<asset type>")
RUNES_VAULT=()
TRUSTED_NON_BTC_RELAYER=""
TRUSTED_FEE_PROVIDER=""
PROTOCOL_FEE_COLLECTOR=""

# gov params
GOV_VOTING_PERIOD="60s"
GOV_EXPEDITED_VOTING_PERIOD="30s"

# Remember to change to other types of keyring like 'file' in-case exposing to outside world,
# otherwise your balance will be wiped quickly
# The keyring test does not require private key to steal tokens from you
KEYRING="test"
#KEYALGO="secp256k1"
KEYALGO="segwit"
LOGLEVEL="info"
# Set dedicated home directory for the $BINARY instance
HOMEDIR="$HOME/testnet"

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

# Reinstall daemon
cd ..
make install
cd scripts

# User prompt if an existing local node configuration is found.
if [ -d "$HOMEDIR" ]; then
	printf "\nAn existing folder at '%s' was found. You can choose to delete this folder and start a new local node with new keys from genesis. When declined, the existing local node is started. \n" "$HOMEDIR"
	echo "Overwrite the existing configuration and start a new local node? [y/n]"
	read -r overwrite
else
	overwrite="Y"
fi

APPHOME="$HOMEDIR/side1"

# Setup local node if overwrite is set to Yes, otherwise skip setup
if [[ $overwrite == "y" || $overwrite == "Y" ]]; then
	# Remove the previous folder
	rm -rf "$HOMEDIR"
	NODE_ID=""
	# Set moniker and chain-id for Cascadia (Moniker can be anything, chain-id must be an integer)
	for ((i=1; i<=N; i++)); do

		APPHOME="${APPHOME%?}$i"
		$BINARY init "$MONIKER #$i" -o --chain-id $CHAINID --home $APPHOME --default-denom "${DENOMS[0]}" &> /dev/null
		
		CONFIG=$APPHOME/config/config.toml
		APP_TOML=$APPHOME/config/app.toml

		sed -i.bak 's/swagger = false/swagger = true/g' $APP_TOML
		sed -i.bak 's/pruning = "default"/pruning = "everything"/g' "$APP_TOML"
		sed -i.bak "s/minimum-gas-prices = \"\"/minimum-gas-prices = \"0.000001${DENOMS[0]}\"/g" $APP_TOML
		sed -i.bak "s/bitcoin_rpc = \"\"/bitcoin_rpc = \"192.248.150.102:18332\"/g" $APP_TOML
		sed -i.bak "s/bitcoin_rpc_user = \"\"/bitcoin_rpc_user = \"side\"/g" $APP_TOML
		sed -i.bak "s/bitcoin_rpc_password = \"\"/bitcoin_rpc_password = \"12345678\"/g" $APP_TOML

		sed -i.bak 's/addr_book_strict = true/addr_book_strict = false/g' $CONFIG
		sed -i.bak 's/allow_duplicate_ip = false/allow_duplicate_ip = true/g' $CONFIG

		if [[ $i -eq 1 ]]; then

			sed -i.bak "s/127.0.0.1:26657/0.0.0.0:26657/g" "$CONFIG"
			sed -i.bak 's/cors_allowed_origins\s*=\s*\[\]/cors_allowed_origins = ["*",]/g' "$CONFIG"

			NODE_ID=$($BINARY comet show-node-id --home $APPHOME)
			for key in "${!DENOMS[@]}"; do
				BALANCES+=",${INITIAL_SUPPLY}${DENOMS[$key]}"
			done
			# echo "$KEY ${BALANCES:1}"
			for ((j=1; j<=N; j++)); do
				$BINARY keys add "v$j" --keyring-backend $KEYRING --algo $KEYALGO --home $APPHOME &> "$HOMEDIR/$j.mnemonic"
				# cp -r $HOMEDIR $APPHOME
				$BINARY genesis add-genesis-account v$j ${BALANCES:1} --keyring-backend $KEYRING --home $APPHOME
			done

			echo "Genesis accounts allocated for local accounts"

			$BINARY genesis gentx v$i ${INITIAL_SUPPLY%?}${DENOMS[0]} --keyring-backend $KEYRING --chain-id $CHAINID --identity "666AC57CC678BEC4" --website="https://side.one" --home $APPHOME

			# set genesis
			GENESIS=$APPHOME/config/genesis.json
			TMP_GENESIS=$APPHOME/config/tmp_genesis.json

			jq --arg max_gas "$MAX_GAS" '.consensus_params["block"]["max_gas"]=$max_gas' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
			jq --arg height "3" '.consensus["params"]["abci"]["vote_extensions_enable_height"]=$height' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
		
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

			# set trusted non btc relayer
			if [ -n "$TRUSTED_NON_BTC_RELAYER" ]; then
				jq --arg relayer "$TRUSTED_NON_BTC_RELAYER" '.app_state["btcbridge"]["params"]["trusted_non_btc_relayers"][0]=$relayer' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
			fi

			# set trusted fee provider
			if [ -n "$TRUSTED_FEE_PROVIDER" ]; then
				jq --arg provider "$TRUSTED_FEE_PROVIDER" '.app_state["btcbridge"]["params"]["trusted_fee_providers"][0]=$provider' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
			fi

			# set protocol fee collector
			if [ -n "$PROTOCOL_FEE_COLLECTOR" ]; then
				jq --arg fee_collector "$PROTOCOL_FEE_COLLECTOR" '.app_state["btcbridge"]["params"]["protocol_fees"]["collector"]=$fee_collector' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
			fi
			
			continue
		fi

		sed -i.bak "s/localhost:9090/localhost:909${i}/g" $APP_TOML

		sed -i.bak "s/127.0.0.1:26657/0.0.0.0:26${i}57/g" $CONFIG
		sed -i.bak "s/127.0.0.1:26658/127.0.0.1:26${i}28/g" $CONFIG
		sed -i.bak "s/0.0.0.0:26656/0.0.0.0:26${i}56/g" $CONFIG
		sed -i.bak "s/localhost:6060/localhost:60${i}0/g" $CONFIG
		sed -i.bak "s/persistent_peers = \"\"/persistent_peers = \"$NODE_ID@localhost:26656\"/g" $CONFIG

		cp -f "${APPHOME%?}1/config/genesis.json" "$APPHOME/config/genesis.json"
		cp -r "${APPHOME%?}1/keyring-$KEYRING" "$APPHOME/keyring-$KEYRING" 

		$BINARY genesis gentx v$i ${INITIAL_SUPPLY%?}${DENOMS[0]} --keyring-backend $KEYRING --chain-id $CHAINID --identity "666AC57CC678BEC4" --website="https://side.one" --home $APPHOME --p2p-port "26${i}56"
		cp -r "$APPHOME/config/gentx/." "${APPHOME%?}1/config/gentx/"

	done

	$BINARY genesis collect-gentxs --home ${APPHOME%?}1 # &> /dev/null
	$BINARY genesis validate --home ${APPHOME%?}1
	# $BINARY keys list --keyring-dir $HOMEDIR --keyring-backend $KEYRING 
	echo "Genesis transactions collected"

	for ((i=2; i<=N; i++)); do
		cp -f "${APPHOME%?}1/config/genesis.json" "${APPHOME%?}$i/config"
		shasum ${APPHOME%?}$i/config/genesis.json
	done

fi

# Cleanup function to run on Ctrl-C
cleanup() {
    # echo "Ctrl-C detected! Killing background jobs..."
    # kill $pid1 $pid2 $pid3
    # wait $pid1 $pid2 $pid3 2>/dev/null
    echo "All background jobs cleaned up. Exiting."
    exit 0
}

# Trap SIGINT (Ctrl-C)
trap cleanup SIGINT

for ((i=START; i<=N; i++)); do
	$BINARY start --home ${APPHOME%?}$i > "$HOMEDIR/output$i.log" 2>&1 &
done

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
$BINARY start --log_level "*:info,oracle:debug" --minimum-gas-prices=0.0001${DENOMS[0]} --home ${APPHOME%?}1
