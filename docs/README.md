# Side Protocol

## 1. Introduction
This document presents a decentralized financial protocol that enables native Bitcoin participation in decentralized finance (DeFi) through advanced cryptographic primitives, primarily leveraging Discreet Log Contracts (DLC) and atomic swap for Bitcoin-Cosmos interoperability. 

## 2. Core Architecture

### 2.1 Bitcoin Compatible Layer
- **BTC Address Compatibility**  
  - Supports native Bech32/Bech32m addresses 
  - Supports mainstream wallet integrations (OKX, Unisat, Ledger)
- **On-chain Bitcoin Light Client**
  - Support SPV (Simple Payment Verfication)
  - 6 Confirmations

### 2.2 Trustless Relayer

Automated cross-chain transaction relay with cryptographic state verification between Bitcoin and Sidechain through:

 - Merkle Inclusion Proof generation/verification
 - Stateless transaction validity checks

**Key Attributes**

| Property | Technical Implementation |
|----------|------------|
|Permissionless|	Open-source client with standardized proof generation rules    |
|Verifiable    |	All relays require SPV (Simplified Payment Verification) proofs|

**Operational Workflow**

- **Bitcoin Layer Monitoring**
  - Continuously scans Bitcoin blocks (post 6-confirmation)
  - Indexes transactions matching predefined vault address patterns

- **Proof Construction**

  Generates compact proofs containing:
  ```protobuf
  message RelayProof {
    bytes tx_hash = 1;  
    bytes merkle_path = 2;  // SHA-256 compression path
    uint32 block_height = 3;  
    bytes coinbase_commitment = 4;  // Block commitment binding
  }
  ```

### 2.3 Decentralized Oracle System

#### 2.3.1 Data Provider Network

  The data provider network comprises all validators (n=100). `Vote Extensions` provide an elegant framework for validators to submit arbitrary off-chain data and achieve on-chain consensus through ABCI++ enhancements. Refer to the [Vote Extensions documentation](https://docs.cosmos.network/main/build/abci/vote-extensions) for implementation details.
  - **BTC Header Synchronization**
    - Implements bitcoin block header synchronization through a Weighted Majority Decision mechanism
    - Requires consensus from validators representing ≥2/3 of total network voting power
  - **Price Feed Mechanism**
    - The price feed aggregates data from the top five cryptocurrency exchanges (Binance, Coinbase, Bybit, OKX, Bitget) utilizing a Time-Weighted Average Price (TWAP) algorithm

#### 2.3.2 Event Signer Network

- **Architecture Overview**  
The Event Signer Network(ESN) implements a decentralized oracle solution for Discreet Log Contracts (DLC) through:  
  - **Threshold Signature Scheme**: FROST network (n=21, k=15) with 21 institutional operators  
  - **Participant Diversity**: Strategically selected entities including:  
    - Major Bitcoin mining pools  
    - Venture capital firms  
    - Infrastructure providers  

- **Security Model**  
While DLC's security fundamentally relies on oracle trustworthiness ([Multi-Oracle Technical Specification](https://github.com/discreetlogcontracts/dlcspecs/blob/master/MultiOracle.md)), our implementation introduces:  

  | Component | Innovation | Security Property |  
  |-----------|-------------|-------------------|  
  | **FROST Framework** | Threshold-optimized Schnorr signatures | Byzantine fault tolerance under m-of-n (15/21) threshold |  
  | **Decentralized Nonce Generation** | Multi-party computation protocol | Bias resistance with ≤1/3 malicious participants |  

  This dual-layer design achieves:  
  - **Trust Minimization**: No single entity controls oracle outputs  
  - **Operational Resilience**: Graceful degradation with 6-node failure tolerance  
  - **Cost Efficiency**: O(n) communication complexity in signing operations

- **Event Announcement**  
  - **Price Event**  
    Price events generate continuous numeric outcomes that cannot be discretely enumerated. This requires [Numeric Outcome Compression](https://github.com/...), which encodes BTC prices (base B=100) into base-B digit prefix arrays. The algorithm outputs nested integer arrays (elements ∈ [0, B)), where each subarray defines allowable digit ranges for price resolution. When a price attestation is finalized, the system automatically generates a new event for the next monitoring interval.  

  - **Date Event**  
    ESN maintains a 365-day rolling event calendar aligned with maximum loan tenures. The oracle executes automated daily updates to advance the coverage window, ensuring continuous temporal attestation availability.  

  - **Loan Event**  
    Initialized during loan origination, repayment-triggered events create cryptographic proof bindings between mainchain DLCs and sidechain states. This mechanism enables non-custodial collateral redemption without DCM coordination.  

- **Event Attestation**  
  Upon event maturity, oracles generate threshold signatures attesting to canonical outcomes. These cryptographically signed attestations are published on-chain through Merkle-root anchoring, enabling DLC participants to deterministically unlock Conditional Execution Transactions (CETs) through adaptor signature decryption.  

- **Event structure**
    ```go
    type OracleEvent struct {
      Type string
      Outcome string 
      OraclePubkey  string
      Nonce   string
      Signature string
      TriggerTime int64
    }
    ```

### 3. Bitcoin Collateralized Lending

#### 3.1 Cryptographic Primitives
- **Discreet Log Contracts (DLC)**  
  - Oracle-signed attestations for price conditions
  - Oracle-signed attestations for date conditions

- **Adaptor Signatures**  
  - Encrypted signature format: σ' = (s', R, T)
  - Key derivation: t = H(R||T)

#### 3.2 Lending Flow Implementation

1. **Vault Creation**  
   - 2-of-2 MuSig address (Borrower + DCM)
   - Collateral ratio: 70% LTV

2. **DLC Contract Lifecycle**  
   - Pre-signed transactions:
     - Settlement CET: Borrower signs 2 outcomes (price_up, price_down)
     - Repayment CET: DCM signs refund path
   - TimeLock enforcement via CHECKSEQUENCEVERIFY

3. **Liquidation Mechanism**  
   - Margin call triggered when:
     - Price < Maintenance Margin (120% LTV)
     - Time expiration
   - Collateral auction via Dutch auction model

### 4. Trust-Minimized Bitcoin Bridge

#### 4.1 Light Client Verification
- SPV proofs validated through:
  - Bitcoin block header Merkle proofs
  - Transaction inclusion verification via:
    ```rust
    fn verify_spv(tx: Transaction, header: BlockHeader, proof: MerkleProof) -> bool {
      // Implementation logic
    }
    ```

#### 4.2 Frost Protocol Vault
- 15-of-21 multisig using FROST (Flexible Round-Optimized Schnorr Threshold)  
  - Distributed key generation
  - 2 round signature aggregation protocols
  - Vault withdrawal limits:
    - 1 BTC daily threshold
    - Emergency freeze via governance proposal

#### 4.3 Process Flow  
  - **Peg In (Cross-chain Deposit)**  
    - User sends BTC from a self-custodied wallet (CEX withdrawals are not supported) to the designated vault address 
    - Relayers continuously scan the Bitcoin mempool, filtering transactions based on predefined vault address patterns  
    - Relayers generate cryptographic Merkle inclusion proofs and submit both raw transaction and proofs to the Sidechain Bridge module 
    - **After verifying 6-block confirmations and proof validity**, the bridge issues corresponding wrapped tokens (1:1 sat = 10^8:1 sBTC) to the originator's sidechain address

  - **Peg Out (Cross-chain Withdrawal)**  
    - Initiated by users transferring sBTC to the bridge module 
    - Bridge automatically constructs a Partially Signed Bitcoin Transaction (PSBT) with vault withdrawal parameters 
    - FROST threshold network monitors the sidechain state for pending withdrawal requests, verifies multisig authorization thresholds  
    - Authorized signers collectively sign the PSBT through distributed key generation  
    - Relayers broadcast finalized PSBTs to the Bitcoin network 
    - Upon Bitcoin network confirmation, the bridge executes sBTC burning corresponding to processed withdrawals

## 5. Security Model

### 5.1 Trust Assumptions
- ≥2/3 honest validators for consensus safety
- FROST protocol security under active adversaries
- DCM module correctness (formal verification completed)
- Bitcoin network finality (6 confirmations)

### 5.2 Risk Analysis & Mitigations

| Risk Category | Mitigation Strategy |
|---------------|---------------------|
| Oracle Delay  | Time-weighted proofs with 3-block confirmation |
| Price Manipulation | 3-phase data aggregation with Coinbase/Binance/Kraken feeds |
| Bridge Attack | FROST signature rotation every 100 blocks |
| DLC Contract Failure | Formal verification using TLA+ models |
| Liquidation Frontrunning | MEV-resistant auction design with privacy pools |

## 6. Conclusion
This protocol establishes a Bitcoin-native DeFi primitive combining Cosmos IBC's interoperability with advanced cryptographic schemes, maintaining Bitcoin's security guarantees while enabling novel financial applications.

---

This document can be expanded with:
1. Cryptographic proof appendix
2. Economic model details
3. Benchmarking data
4. Formal verification reports

Would you like me to elaborate on any specific section?