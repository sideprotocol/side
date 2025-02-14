KEYRING="test"
BINARY="$HOME/go/bin/sided"
CHAINID="devnet"
SENDER="$($BINARY keys show validator -a --keyring-backend $KEYRING )"

echo '${SENDER}'

TSS_BIN="$HOME/workspace/tssigner/target/debug/shuttler"

N=3
T=2

for i in $(seq 1 $N); do
  $TSS_BIN --home $HOME/.shuttler$i init
done

declare -a TSS_PARTICIPANTS
for i in $(seq 1 $N); do
    key=$(jq .pub_key.value $HOME/.shuttler$i/priv_validator_key.json)
    TSS_PARTICIPANTS+=($key)
done

PARTICIPANTS=$(printf "%s," ${TSS_PARTICIPANTS[@]})

ORACLE_PROPOSAL="{
 “messages”: [
 {
  “@type”: “/side.dlc.MsgCreateOracle”,
  “authority”: \""$SENDER"\",
  “participants”: [ ${PARTICIPANTS%?} ],
  “threshold”: $T
 }
 ],
 “metadata”: “”,
 “deposit”: “10000000uside”,
 “title”: “Initiate DKG for oracle”,
 “summary”: “Initiate DKG for oracle”,
 “expedited”: false
}"

echo $ORACLE_PROPOSAL > ../build/oracle.json

DCA_PROPOSAL="{
 “messages”: [
 {
  “@type”: “/side.dlc.MsgCreateAgency”,
  “authority”: \"$SENDER\",
  “participants”: [ ${PARTICIPANTS%?} ],
  “threshold”: $T
 }
 ],
 “metadata”: “”,
 “deposit”: “10000000uside”,
 “title”: “Initiate DKG for DCA”,
 “summary”: “Initiate DKG for DCA”,
 “expedited”: false
}"

echo $DCA_PROPOSAL > ../build/dca.json

$BINARY tx gov submit-proposal ../build/oracle.json --from validator --fees 1000uside --auto gas --chain-id $CHAINID
$BINARY tx gov submit-proposal ../build/dca.json --from validator --fees 1000uside --auto gas --chain-id $CHAINID

sleep 10

$BINARY tx gov vote 1 yes --from validator --fees 1000uside --auto gas --chain-id $CHAINID
$BINARY tx gov vote 2 yes --from validator --fees 1000uside --auto gas --chain-id $CHAINID

for i in $(seq 1 $N); do
  $TSS_BIN --home $HOME/.shuttler$i start > /dev/null &
done