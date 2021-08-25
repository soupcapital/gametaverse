import { BigInt } from "@graphprotocol/graph-ts"
import {
  CryptoBlades,
  FightOutcome,
  InGameOnlyFundsGiven,
  MintWeaponsFailure,
  MintWeaponsSuccess,
  RoleAdminChanged,
  RoleGranted,
  RoleRevoked
} from "../generated/CryptoBlades/CryptoBlades"

import { 
  FightInfo,
  MintWeaponsInfo,
  InGameFundsGivenInfo,
} from "../generated/schema"

export function handleFightOutcome(event: FightOutcome): void {
  // // Entities can be loaded from the store using a string ID; this ID
  // // needs to be unique across all entities of the same type
  // let entity = FightInfo.load(event.transaction.from.toHex())

  // // Entities only exist after they have been saved to the store;
  // // `null` checks allow to create entities on demand
  // if (entity == null) {
  //   entity = new FightInfo(event.transaction.from.toHex())
  // }

  let entity = new FightInfo(event.transaction.from.toHex())
  // Entity fields can be set based on event parameters
  entity.owner = event.params.owner
  entity.character = event.params.character
  entity.weapon = event.params.weapon
  entity.target = event.params.target
  entity.playerRoll = BigInt.fromI32( event.params.playerRoll)
  entity.enemyRoll = BigInt.fromI32(event.params.enemyRoll)
  entity.xpGain = BigInt.fromI32( event.params.xpGain)
  entity.skillGain = event.params.skillGain
  // Entities can be written to the store with `.save()`
  entity.save()

  // Note: If a handler doesn't require existing field values, it is faster
  // _not_ to load the entity from the store. Instead, create it fresh with
  // `new Entity(...)`, set the fields that should be updated and save the
  // entity back to the store. Fields that were not set or unset remain
  // unchanged, allowing for partial updates to be applied.

  // It is also possible to access smart contracts from mappings. For
  // example, the contract that has emitted the event can be connected to
  // with:
  //
  // let contract = Contract.bind(event.address)
  //
  // The following functions can then be called on this contract to access
  // state variables and other data:
  //
  // - contract.DEFAULT_ADMIN_ROLE(...)
  // - contract.GAME_ADMIN(...)
  // - contract.MINT_PAYMENT_RECLAIM_MINIMUM_WAIT_TIME(...)
  // - contract.MINT_PAYMENT_TIMEOUT(...)
  // - contract.PAYMENT_USING_STAKED_SKILL_COST_AFTER_DISCOUNT(...)
  // - contract.blacksmith(...)
  // - contract.burnWeaponFee(...)
  // - contract.characters(...)
  // - contract.fightRewardBaseline(...)
  // - contract.fightRewardGasOffset(...)
  // - contract.fightTraitBonus(...)
  // - contract.fightXpGain(...)
  // - contract.getRoleAdmin(...)
  // - contract.getRoleMember(...)
  // - contract.getRoleMemberCount(...)
  // - contract.hasRole(...)
  // - contract.inGameOnlyFunds(...)
  // - contract.mintCharacterFee(...)
  // - contract.mintWeaponFee(...)
  // - contract.oneFrac(...)
  // - contract.priceOracleSkillPerUsd(...)
  // - contract.promos(...)
  // - contract.randoms(...)
  // - contract.refillStaminaFee(...)
  // - contract.reforgeWeaponFee(...)
  // - contract.reforgeWeaponWithDustFee(...)
  // - contract.skillToken(...)
  // - contract.stakeFromGameImpl(...)
  // - contract.totalInGameOnlyFunds(...)
  // - contract.totalMintPaymentSkillRefundable(...)
  // - contract.weapons(...)
  // - contract.REWARDS_CLAIM_TAX_MAX(...)
  // - contract.REWARDS_CLAIM_TAX_DURATION(...)
  // - contract.getSkillToSubtractSingle(...)
  // - contract.getSkillToSubtract(...)
  // - contract.getSkillNeededFromUserWallet(...)
  // - contract.getMyCharacters(...)
  // - contract.getMyWeapons(...)
  // - contract.unpackFightData(...)
  // - contract.getMonsterPower(...)
  // - contract.getPlayerPower(...)
  // - contract.getPlayerTraitBonusAgainst(...)
  // - contract.getTargets(...)
  // - contract.isTraitEffectiveAgainst(...)
  // - contract.usdToSkill(...)
  // - contract.getTokenRewards(...)
  // - contract.getXpRewards(...)
  // - contract.getTokenRewardsFor(...)
  // - contract.getTotalSkillOwnedBy(...)
  // - contract.getOwnRewardsClaimTax(...)
}

export function handleInGameOnlyFundsGiven(event: InGameOnlyFundsGiven): void {
  let entity = new InGameFundsGivenInfo(event.transaction.from.toHex())
  // Entity fields can be set based on event parameters
  entity.to = event.params.to
  entity.skillAmount = event.params.skillAmount
  // Entities can be written to the store with `.save()`
  entity.save()
}

export function handleMintWeaponsFailure(event: MintWeaponsFailure): void {
  let entity = new MintWeaponsInfo(event.transaction.from.toHex())
  // Entity fields can be set based on event parameters
  entity.minter = event.params.minter
  entity.count = event.params.count
  entity.result = false
  // Entities can be written to the store with `.save()`
  entity.save()
}

export function handleMintWeaponsSuccess(event: MintWeaponsSuccess): void {
  let entity = new MintWeaponsInfo(event.transaction.from.toHex())
  // Entity fields can be set based on event parameters
  entity.minter = event.params.minter
  entity.count = event.params.count
  entity.result = true
  // Entities can be written to the store with `.save()`
  entity.save()
}

export function handleRoleAdminChanged(event: RoleAdminChanged): void {}

export function handleRoleGranted(event: RoleGranted): void {}

export function handleRoleRevoked(event: RoleRevoked): void {}
