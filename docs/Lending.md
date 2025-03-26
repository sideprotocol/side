# Side Finance Litepaper

# 1. Abstract

Side Finance is a decentralized financial system designed to enhance the capital efficiency and utilization of BTC by offering a non-custodial, liquidity pool-based lending solution that does not require third-party custody of the native BTC used as collateral.

# 2. Preliminary

This litepaper provides technical details, but the core principles of the Side Finance lending protocol are straightforward:

- **Not your keys, not your coins:** While BTC can be used as collateral to borrow other assets, the collateralized BTC cannot be arbitrarily spent by lenders under any circumstances unless liquidation occurs. The protocol enforces this principle by securing all BTC collateral in 2-of-2 multi-signature addresses, which require the borrower’s signature for any transaction. This ensures that the protocol offers a Bitcoin-native level of security to its users.
- **Alt-chain risk remains with alt-chain participants:** No crypto loan system can entirely eliminate risk, especially when loan assets involve trusted entities, such as stablecoins (e.g., USDT). However, the protocol is designed to minimize economic risks for liquidity providers, ensuring a trust-minimized environment where risks are managed within the alt-chain ecosystem without adversely affecting borrowers.

Side Finance leverages established Bitcoin security technologies, including Schnorr-based Adaptor Signatures, HTLCs, Taproot, and threshold signing, integrated with Discreet Log Contracts (DLCs) and secure distributed oracles. The combination of these technologies into Scriptless Scripts enables Side Finance to provide smart contract-like functionality on Bitcoin today, without requiring any changes to the underlying Bitcoin protocol’s opcodes.

### 2.1 Threshold Adaptor Signatures

Adaptor signatures [[1](https://www.notion.so/Side-Finance-Litepaper-1058bd3c59fd80e296f4fb76072db64f?pvs=21)] represent a significant cryptographic advancement that enhances the efficiency and privacy of conditional transactions within the Bitcoin network. These signatures serve as a bridge between standard signatures and concealed values, offering dual functionality:

1. They disclose a secret value when combined with the corresponding signature.
2. They generate the complete signature when presented with the secret value.

A notable feature of adaptor signatures is their reusability: third parties can create secondary adaptors from the initial commitment, even without knowledge of the secret value. This characteristic renders them particularly effective for establishing conditional locks in Bitcoin contracts.

Traditionally, Bitcoin contracts employ hashlocks for conditional payments, ensuring atomicity across multiple transactions. While effective, hashlocks have certain limitations, including their on-chain storage footprint and the ability to link transactions that share the same hash across different blockchains.

Adaptor signatures provide developers with greater flexibility, enabling the creation of more intricate conditional structures without increasing on-chain data requirements.

### 2.2 Discreet Log Contracts (DLCs)

At the heart of Side Finance’s infrastructure is the Discreet Log Contract (DLC) [[2](https://www.notion.so/Side-Finance-Litepaper-1058bd3c59fd80e296f4fb76072db64f?pvs=21)], a sophisticated, oracle-based Bitcoin smart contract framework. DLCs allow for conditional payments based on off-chain events, all without involving a third-party custodian to manage funds.

These contracts leverage multiple cryptographic methods, including multi-signature transactions, Hash Time-Locked Contracts (HTLCs) [[3](https://www.notion.so/Side-Finance-Litepaper-1058bd3c59fd80e296f4fb76072db64f?pvs=21)], Schnorr signatures [[4](https://www.notion.so/Side-Finance-Litepaper-1058bd3c59fd80e296f4fb76072db64f?pvs=21)], and adaptor signatures, ensuring the security and non-custodial nature of the system.

A DLC begins with a funding transaction that locks funds from two parties into a 2-of-2 multisignature output. Pre-signed Contract Execution Transactions (CETs) [[5](https://www.notion.so/Side-Finance-Litepaper-1058bd3c59fd80e296f4fb76072db64f?pvs=21)] are created for different potential outcomes of the external event, but these CETs remain inactive until a real event outcome occurs.

The oracle, which is independent of the contract parties, publishes a public key or nonce at the contract’s initiation. Participants use this key to generate adaptor signatures for the CETs. When the oracle reveals the outcome by disclosing a signature, the participants can finalize the relevant CET and broaDCMst it to the Bitcoin network, redistributing the locked funds according to the contract’s terms.

DLCs provide a secure, trustless method to execute conditional payments and smart contract logic, pushing Bitcoin’s capabilities for DeFi applications.

# 3. Side Finance Architecture

## 3.1 Definitions

### 3.1.1 Participants

**Borrower:** The party receiving the loan by using BTC as Collateral

**Liquidity Provider:** Users supplying the loan through the Lending Contract

**DCM (Distributed Collateral Manager):**: A decentralized network of operators organized into a threshold adaptor signature scheme. DCM operators sign Bitcoin 2-of-2 multi-sig transactions on behalf of the Lending Contract, with the counterparty being the Borrower. Additionally, the DCM manages Liquidated Assets in the event of a liquidation.

**Lending Contract:** A smart contract deployed on the Side Chain that automates the operations of the lending pool. This contract enables liquidity providers to supply assets for lending and earn rewards in return. While the Lending Contract itself does not have the capability to sign transactions, it delegates this function to the DCM, which signs transactions on behalf of the contract and in collaboration with the borrower.

**Validator:** The Side Chain is based on [CometBFT](https://docs.cometbft.com/v0.38/) that relies on a set of validators that are responsible for committing new blocks in the blockchain. These validators participate in the consensus protocol by broadcasting votes that contain cryptographic signatures signed by each validator's private key. These validators also provides off-chain data such as BTC price and Bitcoin Headers through [Vote Extension](https://docs.cosmos.network/v0.52/build/abci/vote-extensions)

**Oracle Event Signer:** Monitor the price of BTC from cryptographically signed upstream sources, signing attestations at predetermined intervals. To maintain system integrity, Side Finance implements a mechanism that discards prices falling outside a pre-established variance level. This approach ensures that in the case of problems with a specific oracle, the safety of the system is maintained. Side Finance utilizes multiple independent cryptographically signed price sources to provide outcome attestations for its DLCs. 

### 3.1.2 Glossary

**Collateral Vault:** A Bitcoin Taproot address where borrowers send their BTC as collateral for lending. Each loan has its own unique vault address designated for its collateral.

**Repayment Escrow:** An escrow vault where the borrower locks their loan repayment on the Side Chain

**Collateral:** The BTC assets pledged by the Borrower to borrow the loan

**Principal**: The original amount of assets borrowed

**Maturity Time:** The deadline by which the Borrower must repay the loan in full. If the Borrower fails to repay by this date, the DCM may liquidate the Collateral to recover the outstanding debt

**Liquidated Assets:** Collateral liquidated and sent to the DCM due to the borrower's failure to fulfill payments on the principal and interest of the loan before Maturity Time

**Loan Default**: The failure to repay a loan by the Maturity Time, which can result in the liquidation of Collateral

**Liquidation Price:** The BTC price at which a loan becomes undercollateralized, triggering the execution of CETs

**Final Timeout:** The specified Bitcoin blockchain height after which the Borrower is entitled to reclaim the collateral if the DCM or Side Chain becomes unresponsive.

## 3.2 Workflow Overview

Side Finance defines a liquidity pool-based lending protocol that provides loans to BTC holders and generates returns for loan providers. The loan assets are pooled in smart contracts on the Side Chain and offered to borrowers without involving any third party holding the collateral during the loan period. At a high level, the protocol proceeds as follows:

1. **Loan Assignment**
    1. The borrower sends a loan request to the Lending Contract, including the Maturity Time to initiate the loan. The Lending Contract then assigns a Collateral Vault that is specifically created for the Borrower to lock their collateral.
    2. The Lending Contract computes the interest rate and liquidation price based on the current market price, and pre-allocates the loan amount from available liquidity for the Borrower.
    3. To claim the loan, the borrower deposits the BTC collateral into the assigned Collateral Vault and submits two pre-signed Contract Execution Transactions (CETs) to the Lending Contract. One of the CETs will be executed in the event of liquidation triggered by the depreciation of the BTC collateral value or loan maturity date attainment.
    4. The Distributed Collateral Manager (DCM) additionally pre-signs a third CET and deposited into the Lending Contract, which is programmatically executed upon successful repayment attestation to finalize debt closure and collateral release.
    5. The Lending Contract verifies the validity of the adaptor signatures, If the adaptor signatures are valid the borrower will receive the pre-allocated loan amount. The BTC collateral remains securely locked in the Collateral Vault until the loan reaches the `Maturity Time`.

2. **Repayment**  

3. **Liquidation**
    1. see Section 3.5
4. **Final Timeout**
    1. A `final_timeout` protects the Borrower in the event that the Lending Contract or Side Chain becomes unresponsive.

In the following sections, we will provide more details of each step involved in the lending process as outlined above.

## 3.3 Loan Assignment

The Borrower sends a loan request to the Lending Contract to initiate a loan.

```json
{
	"borrower":"bc1p5d7rjq7g6rdk2yhzks9smlaqtedr4dekq08ge8ztwac72sfr9rusxg3297",
	"maturityTime": "2024-10-10T00:00:00z",
}
```

The Lending Contract then assigns a unique vault, the Collateral Vault, for the loan. Expressed in a Miniscript-like pseudocode, the lock script for this vault is as follows:

```
or(
    // case 1 - the DCM and Borrower can collaborate to create CETs
    and(
        pk(borrower),
        thresh(T, pk(DCM0),pk(DCM1),...pk(DCMn))
    ),
    
    // case 2 - collateral reverts to Borrower after Final Timeout(15 days after Maturity Date), 
    and(
        pk(borrower),
        after(final_timeout)
    )
)
```

The Collateral Vault serves as the foundation for all collateral-related operations. In this paper, we will examine each spending condition in detail. For now, here are some initial observations to help build understanding.

The DCM multi-sig participants are indexed from `0..n`, with a threshold `T` of signatures required to spend. 

The aggregated adaptor signature setup in the script above merits further examination. We can formally describe the Adaptor Signature process as follows. Let `λ` represent the secret scalar, and `G` be the base point of the elliptic curve. The adaptor point `A` is computed as `A = λG`.  Let `m` denote the message, which is pre-signed signature of bitcoin transaction: `m = signature_of_txs`

Given a private key `sk`, a nonce seed `η`, and the adaptor point `A`, we define the **adaptor signature** `σ_A = SignAdaptor(sk, m, η, A)`.

The function `SignAdaptor` represents the cryptographic operation to create an adaptor signature. To redeem the signature, we use the secret `λ` to adapt `σ_A`, producing the **final signature** `σ = Adapt(σ_A, λ)`.

In this formulation:

- `λ ∈ Zq` (the scalar field of the curve)
- `A, G ∈ E(Fq)` (points on the elliptic curve)
- `m ∈ {0,1}^256` (256-bit hash output)
- `σ_A`, `σ` are elements of the signature space

In summary, from the Bitcoin chain’s perspective, the resulting **final signature** appears as a standard Schnorr signature. However, this scheme modifies the signing process to produce a special adaptor signature, which can embed a secret to be revealed at a later stage. This adaptor signature enables trustless lending by leveraging *Scriptless Scripts* within the Bitcoin framework.

The Lending Contract also needs to compute the interest and liquidation price based on the current price and pre-allocate funds for this loan, representing it as the Loan Request:

```json
{
	"borrower":"bc1p5d7rjq7g6rdk2yhzks9smlaqtedr4dekq08ge8ztwac72sfr9rusxg3297",
	"maturityTime": "2025-10-10T00:00:00z",
    "maturityEvent": "9f33f577811da09fc90ad46586308e8c75acef9201fb1a66c",
	"borrowAmount": "20000",
	"vaultAddress": "bc1p6r4u4qlajya6feu337gtngksgeyf4nf0mf4q35gtcj5v0ja9m00q3eagh3",
	"Oracle": "86308e8c75acef9201fb1a66c9f33f577811da09fc90ad465",
	"currentPrice": "50000.00",
    "liquidate_price": "30000.00",
    "liquidate_price_event": "6586308e8c75acef9201fb1a66c9f33f577811da09fc90ad4",
	"collateralAmount": "1",
	"createAt": "2024-10-10T10:10:10z"
}
```

If the Borrower agrees to the terms, they can send the BTC collateral to the Collateral Vault. Once confirmed on Side Chain, the Borrower outputs and signs CETs for a Bitcoin Discreet Log Contract, exchanging them for counter-signing with the DCM. These CETS are:
- Default Liquidation CET
- Price Liquidation CET
- Repayment CET 

In code, the adapter signature generation process is as follows: 

```rust
let adaptor_point = secret.base_point_mul();
let message = sha265(lr_json);
let adaptor_signature = sign_adaptor(seckey, message, nonce_seed, adaptor_point);
let redeem_signature = adaptor_signature.adapt(b"loan_secret");
```

The Borrower then submits the CETs and adaptor signatures to the Lending Contract on the Side Chain to claim loan assets, such as USDC.

At this stage, the collateral is securely locked in the Collateral Vault, and the Borrower has received the loan. The Borrower may repay the loan at any time before the Maturity Date.

## 3.4 Repayment

The borrower must repay the loan before the Maturity Time to avoid liquidation by the DCM.

Recall that during Loan Assignment, BTC collateral for the loan was locked into the Collateral Vault. One of the spending conditions are a 2-of-2 multi-sig UTXO where one side is the DCM, and the other side is the Borrower, with an aggregatable adaptor signature scheme, allowing a adaptor signature from the DCM to publicly reveal a real signature.

Assuming the loan is denominated in USDC, the borrower must submit a transaction to the Lending Contract on the Side Chain that includes:

1. A USDC transfer from the Borrower to the Lending Contract to repay the loan principal and interest.
2. Oracle sign a repayment attestation of a event triggered by Lending contract.
3. Reveal adaptor signature and broadcast CET to Bitcoin

Once the signature is broadcast on Bitcoin, the borrower received their collateral back. 

## 3.5 Liquidation

### 3.5.1 Liquidation Scenarios

Liquidation is governed by a Discreet Log Contract (DLC), which is established during the Loan Assignment process. Collateral will be liquidated in two scenarios:

1. **Collateral Value Depreciation:** 
    - Trigger Conditions:
        - Oracle Network provides Schnorr threshold signatures for BTC/USD TWAP every ~6 seconds
        - 
    - Execution Protocol:
        a. Distributed Collateral Manager (DCM) verifies:
            - 15/21 Oracle signatures validity
            -  Attestation timestamp within 6-block confirmation window
        b. CET adaptor signature decryption via:
The Oracle provides signed attestations of BTC’s price at predefined intervals. The DCM uses the Oracle’s attestation signature to unlock the pre-signed CET and its adaptor signature, allowing the DCM to spend the BTC collateral according to the terms of the CET. This action initiates the liquidation process, with the collateral sent to the DCM for liquidation. The DCM uses the oracle attestation's signature to unlock a pre-signed CET's adaptor signature and spend the BTC collateral according to the CET's terms. This action triggers the loan collateral's liquidation. Collateral is sent to the DCM for liquidation.
2. **Loan Default:** The Oracle regularly provides a signed attestation of the date at UTC 00:00. all signature of CETs corresponsed upon on this date will revealed, therefore the collateral asset will be send from vault to DCM. 

### 3.5.2 Liquidation Flow

The liquidation flow involves selling liquidated assets to recover as much of the loan value as possible. The DCM plays a critical role in mitigating risks for Liquidity Providers during liquidation.

The DCM performs the following tasks:

1. **Receives Liquidated Assets:** When a liquidation is triggered, the DCM takes custody of the liquidated assets.
2. **Collateral Market:** The Lending Contract lists in the collateral market. Everyone can buy these collateral asset at current price with additional 5% bonus.
3. **Repaying the Pool:** The auction proceeds, along with a liquidity penalty, are sent to the Lending Contract to cover the principal and interest.
4. **Surplus or Deficit:** If there is a surplus from the auction, it is returned to the borrower to minimize the impact of liquidation. However, if the auction proceeds are insufficient to cover the loan principal, the liquidity providers incur the loss.

## 3.6 Loss of Liveness

A `final_timeout` timelock in the Collateral Vault safeguards the Borrower in the event that the DCM or the Side Chain becomes unresponsive. In such cases, all locked BTC collateral is returned to the Borrower, mitigating liveness risks associated with non-Bitcoin components.

The `final_timeout` must occur after the loan’s Maturity Time.

# 4. Side Chain

Side Finance is a non-custodial lending solution for Bitcoin, designed as a cross-chain lending system. In this framework, BTC collateral is securely locked on the Bitcoin network without the need for third-party custody, while loans are issued on Lending Contract which is deployed on a separate distributed ledger, Side Chain.

Side Chain is an independent distributed ledger operated and governed by validators. It hosts the lending contract but does not hold or control the BTC collateral involved in the Side Finance protocol. Consequently, the security of Side Finance relies primarily on the trust assumptions inherent to the Bitcoin network, rather than those of Side Chain. In case of liveness failure of Side Chain, Borrowers may reclaim Collateral after `final_timeout`.

Side Chain leverages CometBFT, a high-performance consensus engine that serves as its backbone. This architecture facilitates fast transaction finality and high throughput, making it ideal for applications that require quick confirmation times. Smart contracts on Side Chain are executed in a Wasm VM, written in Rust for performance and security advantages. Rust's memory safety and Wasm's compatibility reduce vulnerabilities like re-entrancy attacks, common in Ethereum-based smart contracts.

This alt-chain system is specifically crafted with Bitcoin-centric features to provide a seamless and frictionless user experience, and it paves the way for further innovations in decentralized finance for Bitcoin:

- **Bitcoin Address Compatibility:** Side Chain is fully compatible with Bitcoin addresses, allowing users to interact with it without the need to create new wallets or manage multiple addresses.
- **Bitcoin Wallet Compatibility:** Users can access Side Chain using their existing Bitcoin wallets, simplifying the user experience and expanding access to the appchain by leveraging the Bitcoin wallet ecosystem.
- **BTC as Native Gas Token:** Unlike other blockchains that use their own native tokens for gas fees, Side Chain utilizes BTC as its native gas token.
- **BTC Bridging:** Side Bridge functions as the primary bridge for transferring native BTC and Bitcoin-based assets (like Runes) between the Side Chain and the Bitcoin network. The initial version of the bridge uses threshold signatures, making it custodial but suitable for users with a higher tolerance for trust. This bridging feature operates independently and is not integrated into the Side Finance design.
- **DeFi Services:** Side Chain, as a Bitcoin appchain, offers a range of decentralized financial services, including a decentralized exchange embedded within the system, allowing users to seamlessly trade BTC, Bitcoin assets, and other crypto assets.
- **Interoperability:** Side Chain integrates with several established cross-chain communication protocols, facilitating the bridging of assets from various blockchain systems. These assets can be injected as liquidity into Side Finance and utilized within other financial services on Side Chain.

Side Chain is a sub-protocol within the Side Protocol stack. The overview provided above offers a high-level summary. A more detailed technical review can be found in a separate document.

# 5. Security

Side Finance operates as a non-custodial solution, meaning no third party holds the native BTC on the Bitcoin blockchain throughout the loan's duration. The main trust assumption arises during two key processes: the liquidation of collateral through auctions and the repayment settlement.

## 5.1 BTC Collateral Security

The Collateral Vault is secured using a 2-of-2 multi-sig, ensuring that BTC collateral cannot be moved without the Borrower’s authorization. There are four methods to spend UTXOs associated with this address, all of which adhere to Bitcoin-native security principles:

1. using CETs of a Discreet Log Contract (DLC) in the event of collateral value depreciation, relying on Schnorr adaptor signatures
2. a final hash timelock that allows the Borrower to reclaim all collateral in the event that Side Finance stops responding

In no case does any single entity have discretion over a Borrower’s BTC collateral outside the terms specified in the “contract” encoded within the DLC.

## 5.2 Loan Asset Security

The Lending Contract, a smart contract deployed on the Side Chain, governs the management of loan assets according to predefined rules encoded within its framework. Liquidity Providers have the flexibility to enter or exit the liquidity pool at their discretion.

## 5.3 Oracle Security

In a Bitcoin DLC-based DeFi system like Side Protocol, oracles must cryptographically sign periodic price attestations, enabling the liquidation of BTC when necessary. Oracle operations need to be decentralized, run by operators who have an economic stake in the system that can be slashed if they act maliciously.

To reducing risk, we divide oracle into two parts: Data Provider and Event Signer, 

### 5.3.1 Data Provider Security

Data provider are all validators, who fetch prices from top exchanges and aggregated with TWAP algorithm and sync Bitcoin header through a Weighted Majority Decision mechanism, 2/3 of total voting prower is required to submited a new price on chain.

So every valdiator can impace the final result(price or header) according to their voting power.

### 5.3.2 Event Signer Security

Event Signer are frost network, 7 well-known Side Chain validators, chosen through on-chain governance by stakers, also serve as Event Signer. These validators are selected due to their vested interest in the system's success and their role in securing value across multiple chains. Should any validator provide incorrect price outputs, a cryptographically signed proof of misbehavior is generated. This proof automatically affects their stake holdings on Side and damages their reputation on other chains they secure.

It’s important to note that Event Signer cannot benefit directly from any price liquidation events they sign, that they cannot make up arbitrary prices, and that a threshold set of operators must agree on the current price provided by all validators to be signed. not individual signer would able to fraud to the protoccol.

The Lending contract collectively produce a stream of DLC announcements for potential future price events. Later, they sign an attestation of the real BTCUSD price for each time period as it occurs. These DLC announcement streams and attestations are then used by the DCM to liquidate loans that fall below the Liquidation Price.

The Lending contract collectively produce a stream of DLC announcements for future date events. Later, they sign an attestation of the date as it comes. These DLC announcement streams and attestations are then used by the DCM to liquidate loans that is defaulted.

## 5.4 DCM Security

The Distributed Collateral Manager (DCM) acts on behalf of the Lending Contract to sign necessary transactions; however, it’s important to clarify that it doesn’t have control over the Lending Contract or the Collateral Vault. The funds supplied by lenders remain non-custodial and securely held within the smart contract. The DCM’s role is limited to temporarily holding liquidated assets (BTC collateral) for auction purposes. To reduce risk, these liquidated assets must be sold as quickly as possible.

The DCM essentially functions as an agent managing the bad debt of liquidity providers. Allowing liquidity providers to nominate the DCM aligns interests and fosters trust. This nomination and selection process occurs periodically, with voting power weighted by the amount and type of LP tokens held. Crucially, the DCM and the Oracle (managed by 21 Side Chain validators) must remain separate entities.

Any misconduct by DCM operators will lead to penalties, including potential removal from the DCM role.

Additionally, all DCM operators must comply with Know Your Customer (KYC) requirements. This ensures all operators are properly vetted and held accountable for their actions within the network.

## References

[1] Bitcoin Ops *“Adaptor Signatures” -* [https://bitcoinops.org/en/topics/adaptor-signatures/](https://bitcoinops.org/en/topics/adaptor-signatures/) 

[2] Thaddeus Dryja *“Discreet Log Contracts”* - [https://adiabat.github.io/dlc.pdf](https://adiabat.github.io/dlc.pdf)

[3] Bitcoin Wiki *“Hash Time Locked Contracts”* - [https://en.bitcoin.it/wiki/Hash_Time_Locked_Contracts](https://en.bitcoin.it/wiki/Hash_Time_Locked_Contracts)

[4] Victor Shoup *“The Many Faces of Schnorr”* - [https://eprint.iacr.org/2023/1019.pdf](https://eprint.iacr.org/2023/1019.pdf) 

[5] DLC Specs *“Introduction to DLCs”* - [https://github.com/discreetlogcontracts/dlcspecs/blob/master/Introduction.md](https://github.com/discreetlogcontracts/dlcspecs/blob/master/Introduction.md)