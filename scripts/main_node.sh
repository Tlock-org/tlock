#!/bin/bash
# Run this script to quickly install, setup, and run the current version of the network without docker.
#
# Examples:
# CHAIN_ID="localchain-1" HOME_DIR="~/.tlock" BLOCK_TIME="1000ms" CLEAN=true sh scripts/test_node.sh
# CHAIN_ID="localchain-2" HOME_DIR="~/.tlock" CLEAN=true RPC=36657 REST=2317 PROFF=6061 P2P=36656 GRPC=8090 GRPC_WEB=8091 ROSETTA=8081 BLOCK_TIME="500ms" sh scripts/test_node.sh
export KEY="node0"
export KEY1="node1"
export KEY2="node2"
export KEY3="node3"
export KEY4="node4"
export KEY5="node5"
#export U_KEY1="u1"
#export U_KEY2="u2"
#export U_KEY3="u3"
#export U_KEY4="u4"
#export U_KEY5="u5"
#export U_KEY6="u6"
#export U_KEY7="u7"
#export U_KEY8="u8"
#export U_KEY9="u9"
#export U_KEY10="u10"
export CHAIN_ID=${CHAIN_ID:-tlock-mainnet-1}
export MONIKER="tlock"
export KEYALGO="secp256k1"
export KEYRING=${KEYRING:-"file"}
export HOME_DIR=$(eval echo "${HOME_DIR:-"~/.tlock"}")
export BINARY=${BINARY:-tlockd}
export DENOM=${DENOM:-uTOK}
export CLEAN=${CLEAN:-"true"}
export RPC=${RPC:-"26657"}
export REST=${REST:-"1317"}
export PROFF=${PROFF:-"6060"}
export P2P=${P2P:-"26656"}
export GRPC=${GRPC:-"9090"}
export GRPC_WEB=${GRPC_WEB:-"9091"}
export PROFF_LADDER=${PROFF_LADDER:-"6060"}
export ROSETTA=${ROSETTA:-"8080"}
export BLOCK_TIME=${BLOCK_TIME:-"3s"}
# if which binary does not exist, install it
if [ -z `which $BINARY` ]; then
  make install
  if [ -z `which $BINARY` ]; then
    echo "Ensure $BINARY is installed and in your PATH"
    exit 1
  fi
fi
alias BINARY="$BINARY --home=$HOME_DIR"
command -v $BINARY > /dev/null 2>&1 || { echo >&2 "$BINARY command not found. Ensure this is setup / properly installed in your GOPATH (make install)."; exit 1; }
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }
set_config() {
  $BINARY config set client chain-id $CHAIN_ID
  $BINARY config set client keyring-backend $KEYRING
}
set_config
from_scratch () {
  # Fresh install on current branch
  make install
  # remove existing daemon files.
  if [ ${#HOME_DIR} -le 2 ]; then
      echo "HOME_DIR must be more than 2 characters long"
      return
  fi
  rm -rf $HOME_DIR && echo "Removed $HOME_DIR"
  # reset values if not set already after whipe
  set_config
  add_key() {
    key=$1
    mnemonic=$2
    echo $mnemonic | BINARY keys add $key --keyring-backend $KEYRING --algo $KEYALGO --recover
  }
  # === MAINNET KEY MANAGEMENT ===
  # Smart key import with fallback:
  # - Both MAINNET_VALIDATOR_MNEMONIC and MAINNET_VALIDATOR_PASSWORD set: fully automated import
  # - Only MAINNET_VALIDATOR_MNEMONIC set: guided interactive mode (you can paste mnemonic)
  # - Neither set: create new key interactively
  
  echo "=========================================="
  echo "  MAINNET VALIDATOR KEY SETUP"
  echo "=========================================="
  
  # Check if validator key already exists
  if BINARY keys show $KEY --keyring-backend $KEYRING > /dev/null 2>&1; then
    echo "✓ Validator key '$KEY' already exists"
    echo "  Address: $(BINARY keys show $KEY -a --keyring-backend $KEYRING)"
  else
    echo "Validator key '$KEY' not found. Creating..."
    echo ""
    
    # Scenario 1: Both mnemonic and password provided (fully automated)
    if [ -n "$MAINNET_VALIDATOR_MNEMONIC" ] && [ -n "$MAINNET_VALIDATOR_PASSWORD" ]; then
      echo "Mode: Fully automated import (mnemonic + password from environment)"
      echo "Importing validator key..."
      
      # Use printf to provide mnemonic and password non-interactively
      printf "%s\n%s\n%s\n" "$MAINNET_VALIDATOR_MNEMONIC" "$MAINNET_VALIDATOR_PASSWORD" "$MAINNET_VALIDATOR_PASSWORD" | \
        BINARY keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover > /tmp/key_import_result.txt 2>&1
      
      if [ $? -eq 0 ]; then
        echo "✓ Validator key imported successfully"
        rm -f /tmp/key_import_result.txt
      else
        echo "❌ Failed to import key. Error details:"
        cat /tmp/key_import_result.txt
        rm -f /tmp/key_import_result.txt
        exit 1
      fi
    
    # Scenario 2: Only mnemonic provided (guided interactive)
    elif [ -n "$MAINNET_VALIDATOR_MNEMONIC" ]; then
      echo "Mode: Guided import (mnemonic from environment)"
      echo ""
      echo "Your mnemonic is ready in the environment variable."
      echo "You will be prompted to:"
      echo "  1. Enter your mnemonic (paste from: echo \$MAINNET_VALIDATOR_MNEMONIC)"
      echo "  2. Enter keyring password (twice)"
      echo ""
      echo "💡 Tip: Run this command to see your mnemonic:"
      echo "   echo \$MAINNET_VALIDATOR_MNEMONIC"
      echo ""
      read -p "Press ENTER to continue..."
      echo ""
      
      # Interactive mode - user will type/paste everything
      BINARY keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover
      
      if [ $? -eq 0 ]; then
        echo ""
        echo "✓ Validator key imported successfully"
      else
        echo ""
        echo "❌ Failed to import key. Please check your inputs."
        exit 1
      fi
    
    # Scenario 3: No mnemonic provided (fully interactive)
    else
      # Warn if password is set but mnemonic is not
      if [ -n "$MAINNET_VALIDATOR_PASSWORD" ]; then
        echo "⚠️  WARNING: MAINNET_VALIDATOR_PASSWORD is set but MAINNET_VALIDATOR_MNEMONIC is not."
        echo "⚠️  The password will be ignored. A new key will be created interactively."
        echo ""
      fi
      
      echo "Mode: Create new validator key (fully interactive)"
      echo ""
      echo "⚠️  WARNING: A new validator key will be created!"
      echo "⚠️  You MUST backup the mnemonic phrase that will be displayed!"
      echo ""
      read -p "Press ENTER to continue or Ctrl+C to abort..."
      echo ""
      
      BINARY keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO
      
      if [ $? -eq 0 ]; then
        echo ""
        echo "=========================================="
        echo "⚠️  CRITICAL: BACKUP YOUR MNEMONIC NOW!"
        echo "=========================================="
        echo "The mnemonic phrase was displayed above."
        echo "Write it down and store it in a SAFE place."
        echo "Without it, you CANNOT recover your validator!"
        echo ""
        read -p "I have backed up my mnemonic [press ENTER]..."
      else
        echo ""
        echo "❌ Failed to create key."
        exit 1
      fi
    fi
    
    echo ""
    echo "Validator Address: $(BINARY keys show $KEY -a --keyring-backend $KEYRING)"
  fi
  echo "=========================================="
  echo ""

#  # u1: tlock17a8553c94wy9q38ppmetx3f9gl4rp7qh7e5knx
#  add_key $U_KEY1 "evil broccoli pink assume abuse skill clog south minor oblige unfair palm end title cabbage siren unaware gossip deer naive topic minimum weapon social"
#  # u2: tlock1dpe2cmv8uahndy5rz8lkp0fvu46c9uls5kqz4q
#  add_key $U_KEY2 "chat bronze fly siren decline engine concert increase aerobic defense material style dish bar conduct carry lyrics swear example file reunion addict figure question"
#  # u3: tlock19qs3ldj4rvfgv5s73zjh2knj5jscvr9awpct6d
#  add_key $U_KEY3 "boat ranch ask fire welcome toilet cable response success slab happy use annual decorate fame state profit lyrics again leaf ecology cash fault setup"
#  # u4: tlock1zcts2px3h9dhdgvtp8w3uf9du9qnecfl78tfp6
#  add_key $U_KEY4 "cage jazz bring web youth exchange position sail wheel price random vanish speak wheat purpose famous emotion ocean civil space level truck auction situate"
#  # u5: tlock1y037hez2c5d206nktj0ffl87smg0y24tdwl85s
#  add_key $U_KEY5 "furnace orbit train tone fame woman comic soccer palm note access brand parrot vague eye cause barely canal quick wine mom town income risk"
#  # u6: tlock1mh6sc4t7gqhru78ndnw63ngxzsgp7q4yga4emw
#  add_key $U_KEY6 "pencil retreat truly calm forget drip illegal road rather museum tuna riot inch fork old buffalo misery girl erase style remember fee twist wrong"
#  # u7: tlock10c5vxm6crtf0q3yqdt22s44ww8wdptszq0ajr8
#  add_key $U_KEY7 "insane grocery spoon oppose tower charge bubble force abuse cannon pole result magic common spike snack act seminar fruit achieve roast kid pistol neutral"
#  # u8: tlock1k2vsd5tj34gms3jdssru8ksrqx3gmhjmxnm4z4
#  add_key $U_KEY8 "garage human destroy regular crisp crater unlock hole keen ladder rose trial dose zone shiver arena hair gloom raven bonus close wet tourist basic"
#  # u9: tlock143elfmf479sj9x4tzgx50cvspaexwsehedhgts
#  add_key $U_KEY9 "case gossip vibrant trap that slice file velvet click tide mountain hungry lawsuit repair nest jar reopen affair nation true rug advice endless cricket"
#  # u10: tlock140ywnzsvj0np3mvv7cemcve55zp2z3tt4tlgv5
#  add_key $U_KEY10 "flag angry illegal success today north aspect author verb soft prize horn cost style wrestle fabric circle cost valley one thumb easily inform reject"

  # chain initial setup
  BINARY init $MONIKER --chain-id $CHAIN_ID --default-denom $DENOM
  update_test_genesis () {
    cat $HOME_DIR/config/genesis.json | jq "$1" > $HOME_DIR/config/tmp_genesis.json && mv $HOME_DIR/config/tmp_genesis.json $HOME_DIR/config/genesis.json
  }
  # === CORE MODULES ===
  # Block
  update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
#  update_test_genesis '.consensus_params["block"]["max_gas"]="10000000"'
  # Gov
  update_test_genesis `printf '.app_state["gov"]["params"]["min_deposit"]=[{"denom":"%s","amount":"1000000"}]' $DENOM`
  update_test_genesis '.app_state["gov"]["params"]["voting_period"]="30s"'
  update_test_genesis '.app_state["gov"]["params"]["max_deposit_period"]="600s"'
  update_test_genesis '.app_state["gov"]["params"]["expedited_voting_period"]="15s"'
  # staking
  update_test_genesis `printf '.app_state["staking"]["params"]["bond_denom"]="%s"' $DENOM`
  update_test_genesis '.app_state["staking"]["params"]["min_commission_rate"]="0.050000000000000000"'
  # mint
  update_test_genesis `printf '.app_state["mint"]["params"]["mint_denom"]="%s"' $DENOM`
  # crisis
  update_test_genesis `printf '.app_state["crisis"]["constant_fee"]={"denom":"%s","amount":"1000"}' $DENOM`
  # === CUSTOM MODULES ===
  # tokenfactory
  update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_fee"]=[]'
  update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_gas_consume"]=100000'

  # set TOK decimal
  update_test_genesis '.app_state["bank"]["denom_metadata"] = [
      {
        "description": "tlock token",
        "denom_units": [
          { "denom": "uTOK", "exponent": 0 },
          { "denom": "TOK", "exponent": 6 }
        ],
        "base": "uTOK",
        "display": "TOK",
        "name": "TOK",
        "symbol": "TOK"
      }
    ]'

  # Allocate genesis accounts
  echo "Adding genesis account for validator..."
  BINARY genesis add-genesis-account $KEY 100000000000000000$DENOM --keyring-backend $KEYRING --append
  echo "✓ Genesis account added: $(BINARY keys show $KEY -a --keyring-backend $KEYRING)"
  echo ""
#  BINARY genesis add-genesis-account $KEY1 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $KEY2 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $KEY3 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $KEY4 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $KEY5 10000000000000000000$DENOM --keyring-backend $KEYRING --append

#  BINARY genesis add-genesis-account $U_KEY1 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $U_KEY2 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $U_KEY3 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $U_KEY4 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $U_KEY5 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $U_KEY6 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $U_KEY7 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $U_KEY8 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $U_KEY9 10000000000000000000$DENOM --keyring-backend $KEYRING --append
#  BINARY genesis add-genesis-account $U_KEY10 10000000000000000000$DENOM --keyring-backend $KEYRING --append

  # Sign genesis transaction
  echo "Creating genesis transaction (gentx)..."
  echo "Staking 1000000 $DENOM (1 TOK) for validator..."
  BINARY genesis gentx $KEY 1000000$DENOM --keyring-backend $KEYRING --chain-id $CHAIN_ID
  echo "✓ Genesis transaction created"
  echo ""

  echo "Collecting genesis transactions..."
  BINARY genesis collect-gentxs
  echo "✓ Genesis transactions collected"
  echo ""
  
  echo "Validating genesis file..."
  BINARY genesis validate-genesis
  err=$?
  if [ $err -ne 0 ]; then
    echo "❌ Failed to validate genesis"
    return
  fi
  echo "✓ Genesis validated successfully"
  echo ""
  echo "=========================================="
  echo "  MAINNET INITIALIZATION COMPLETE"
  echo "=========================================="
  echo "Chain ID: $CHAIN_ID"
  echo "Validator: $(BINARY keys show $KEY -a --keyring-backend $KEYRING)"
  echo "Home Directory: $HOME_DIR"
  echo "=========================================="
  echo ""
}
#t set to false
if [ "$CLEAN" != "false" ]; then
  echo "Starting from a clean state"
  from_scratch
fi
echo "Starting node..."
echo "=========================================="
echo "  MAINNET NODE STARTING"
echo "=========================================="
echo "Chain ID: $CHAIN_ID"
echo "RPC: http://0.0.0.0:$RPC"
echo "REST API: http://0.0.0.0:$REST"
echo "gRPC: 0.0.0.0:$GRPC"
echo "Home: $HOME_DIR"
echo "=========================================="
echo ""
# Opens the RPC endpoint to outside connections
sed -i -e 's/laddr = "tcp:\/\/127.0.0.1:26657"/c\laddr = "tcp:\/\/0.0.0.0:'$RPC'"/g' $HOME_DIR/config/config.toml
sed -i -e 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["\*"\]/g' $HOME_DIR/config/config.toml
# REST endpoint
sed -i -e 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:'$REST'"/g' $HOME_DIR/config/app.toml
sed -i -e 's/enable = false/enable = true/g' $HOME_DIR/config/app.toml
sed -i -e 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/g' $HOME_DIR/config/app.toml
# open swagger
sed -i -e 's/swagger = false/swagger = true/g' $HOME_DIR/config/app.toml
# peer exchange
sed -i -e 's/pprof_laddr = "localhost:6060"/pprof_laddr = "localhost:'$PROFF'"/g' $HOME_DIR/config/config.toml
sed -i -e 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:'$P2P'"/g' $HOME_DIR/config/config.toml
# GRPC
sed -i -e 's/address = "localhost:9090"/address = "0.0.0.0:'$GRPC'"/g' $HOME_DIR/config/app.toml
sed -i -e 's/address = "localhost:9091"/address = "0.0.0.0:'$GRPC_WEB'"/g' $HOME_DIR/config/app.toml
# Rosetta Api
sed -i -e 's/address = ":8080"/address = "0.0.0.0:'$ROSETTA'"/g' $HOME_DIR/config/app.toml
# Faster blocks
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "'$BLOCK_TIME'"/g' $HOME_DIR/config/config.toml
#sed -i -e 's/timeout_propose = "1s"/timeout_propose = "1s"/g' $HOME_DIR/config/config.toml
#sed -i -e 's/timeout_prevote = "1s"/timeout_prevote = "1s"/g' $HOME_DIR/config/config.toml
#sed -i -e 's/timeout_precommit = "1s"/timeout_precommit = "1s"/g' $HOME_DIR/config/config.toml
# Start the node with 0 gas fees
BINARY start --pruning=nothing  --minimum-gas-prices=1$DENOM --rpc.laddr="tcp://0.0.0.0:$RPC"