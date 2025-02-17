KEYRING="test"
BINARY="$HOME/go/bin/sided"
CHAINID="devnet"
SENDER="$($BINARY keys show validator -a --keyring-backend $KEYRING )"

TSS_BIN="$HOME/workspace/tssigner/target/debug/shuttler"

N=3
T=2

# User prompt if an existing local node configuration is found.
if [ -d "$HOME/.shuttler1" ]; then
	printf "\nAn existing folder at '%s' was found. You can choose to delete this folder and start a new shuttler with new keys. When declined, the existing local node is started. \n" "$HOME/.shuttlerN"
	echo "Overwrite the existing configuration and start a new shuttler cluster? [y/n]"
	read -r overwrite
else
	overwrite="Y"
fi

if [[ $overwrite == "y" || $overwrite == "Y" ]]; then

  rm -rf $HOME/.shuttler*

  # Init shuttler home
  for i in $(seq 1 $N); do
    $TSS_BIN --home $HOME/.shuttler$i init --network testnet --port "535$i"
    sed -i -e "/rpc_address =/ s/= .*/= \"127.0.0.1:818$i\"/" $HOME/.shuttler$i/config.toml
  done

  # fund relayer address
  for i in $(seq 1 $N); do
    echo "\n *** Fund relayer $i ***"
    $BINARY tx bank send validator $($TSS_BIN --home "$HOME/.shuttler$i" address | grep tb1) 10000000uside --chain-id $CHAINID --keyring-backend $KEYRING --yes --fees 2000uside
    sleep 6
  done
fi

declare -a TSS_PARTICIPANTS
for i in $(seq 1 $N); do
    key=$(jq .pub_key.value $HOME/.shuttler$i/priv_validator_key.json)
    TSS_PARTICIPANTS+=($key)
done

PARTICIPANTS=$(printf "%s," ${TSS_PARTICIPANTS[@]})

ORACLE_PROPOSAL="{
 \"messages\": [
 {
  \"@type\": \"/side.dlc.MsgCreateOracle\",
  \"authority\": \"side10d07y265gmmuvt4z0w9aw880jnsr700jwrwlg5\",
  \"participants\": [ ${PARTICIPANTS%?} ],
  \"threshold\": $T
 }
 ],
 \"metadata\": \"\",
 \"deposit\": \"10000000uside\",
 \"title\": \"Initial dkg for oracle\",
 \"summary\": \"Initiate DKG for oracle\",
 \"expedited\": false
}"

echo $ORACLE_PROPOSAL > ../build/oracle.json
cat ../build/oracle.json

DCA_PROPOSAL="{
 \"messages\": [
 {
  \"@type\": \"/side.dlc.MsgCreateAgency\",
  \"authority\": \"side10d07y265gmmuvt4z0w9aw880jnsr700jwrwlg5\",
  \"participants\": [ ${PARTICIPANTS%?} ],
  \"threshold\": $T
 }
 ],
 \"metadata\": \"\",
 \"deposit\": \"10000000uside\",
 \"title\": \"Initial dkg for agency\",
 \"summary\": \"Initiate DKG for agency\",
 \"expedited\": false
}"

$BINARY tx gov submit-proposal ../build/oracle.json --from validator --fees 1000uside --chain-id $CHAINID --keyring-backend $KEYRING -y
sleep 6
echo $DCA_PROPOSAL > ../build/dca.json
cat ../build/dca.json
$BINARY tx gov submit-proposal ../build/dca.json --from validator --fees 1000uside --chain-id $CHAINID --keyring-backend $KEYRING -y

sleep 6

# Vote active proposals
$BINARY q gov proposals --output json | jq -r '.proposals| .[] | select(.status == 2) | .id'| while read -r id; do 
  echo "\n   >>>> Vote for $id"; 
  $BINARY tx gov vote $id yes --from validator --fees 1000uside --chain-id $CHAINID --keyring-backend $KEYRING -y
  sleep 6
done

for i in $(seq 1 $N); do
  # RUST_BACKTRACE=1 
  $TSS_BIN --home $HOME/.shuttler$i start --oracle --agency > $HOME/.shuttler$i/output.log &
done

tail -f $HOME/.shuttler1/output.log
