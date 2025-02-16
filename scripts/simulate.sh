KEYRING="test"
BINARY="$HOME/go/bin/sided"
CHAINID="devnet"
SENDER="$($BINARY keys show validator -a --keyring-backend $KEYRING )"

TSS_BIN="$HOME/workspace/tssigner/target/debug/shuttler"

N=3
T=2

rm -rf $HOME/.shuttler*

for i in $(seq 1 $N); do
  $TSS_BIN --home $HOME/.shuttler$i init --network testnet --port "535$i"
done

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

sleep 10

# Vote active proposals
$BINARY q gov proposals --output json | jq -r '.proposals| .[] | select(.status == 2) | .id'| while read -r id; do 
  echo "\n   >>>> Vote for $id"; 
  $BINARY tx gov vote $id yes --from validator --fees 1000uside --chain-id $CHAINID --keyring-backend $KEYRING -y
  sleep 6
done

for i in $(seq 1 $N); do
  $TSS_BIN --home $HOME/.shuttler$i start > $HOME/.shuttler$i/output.log &
done

tail -f $HOME/.shuttler$i/output.log
