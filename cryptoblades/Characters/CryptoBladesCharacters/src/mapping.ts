import { BigInt } from "@graphprotocol/graph-ts"
import {
  Characters,
  Approval,
  ApprovalForAll,
  LevelUp,
  NewCharacter,
  RoleAdminChanged,
  RoleGranted,
  RoleRevoked,
  Transfer
} from "../generated/Characters/Characters"
import { CharacterInfo } from "../generated/schema"

export function handleApproval(event: Approval): void {}

export function handleApprovalForAll(event: ApprovalForAll): void {}

export function handleLevelUp(event: LevelUp): void {
  // Entities can be loaded from the store using a string ID; this ID
  // needs to be unique across all entities of the same type
  let entity = CharacterInfo.load(event.params.character.toHex())
  if (entity == null) {
    return 
  }
  entity.level = event.params.level

  // Entities can be written to the store with `.save()`
  entity.save()
}

export function handleNewCharacter(event: NewCharacter): void {
  // Entities can be loaded from the store using a string ID; this ID
  // needs to be unique across all entities of the same type
  //let entity = CharacterInfo.load(event.params.character.toHex())

  let contract = Characters.bind(event.address)
  let entity = new CharacterInfo(event.params.character.toHex())
  entity.owner = event.params.minter
  //let (xp, level, trait, head, arms, torso, legs, boots, race)  = contract.get(event.params.character)
  let info = contract.get(event.params.character)
  // BigInt and BigDecimal math are supported
  entity.xp = info.value0 
  entity.level = info.value1
  entity.trait = info.value2
  entity.staminaTimestamp = info.value3
  entity.head = info.value4
  entity.arms = info.value5
  entity.torso = info.value6
  entity.legs = info.value7
  entity.boots = info.value8
  entity.race = info.value9

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
  // - contract.NO_OWNED_LIMIT(...)
  // - contract.RECEIVE_DOES_NOT_SET_TRANSFER_TIMESTAMP(...)
  // - contract.TRANSFER_COOLDOWN(...)
  // - contract.balanceOf(...)
  // - contract.baseURI(...)
  // - contract.characterLimit(...)
  // - contract.getApproved(...)
  // - contract.getRoleAdmin(...)
  // - contract.getRoleMember(...)
  // - contract.getRoleMemberCount(...)
  // - contract.hasRole(...)
  // - contract.isApprovedForAll(...)
  // - contract.lastTransferTimestamp(...)
  // - contract.maxStamina(...)
  // - contract.name(...)
  // - contract.ownerOf(...)
  // - contract.promos(...)
  // - contract.raidsDone(...)
  // - contract.raidsWon(...)
  // - contract.secondsPerStamina(...)
  // - contract.supportsInterface(...)
  // - contract.symbol(...)
  // - contract.tokenByIndex(...)
  // - contract.tokenOfOwnerByIndex(...)
  // - contract.tokenURI(...)
  // - contract.totalSupply(...)
  // - contract.transferCooldownEnd(...)
  // - contract.transferCooldownLeft(...)
  // - contract.get(...)
  // - contract.getLevel(...)
  // - contract.getRequiredXpForNextLevel(...)
  // - contract.getPower(...)
  // - contract.getPowerAtLevel(...)
  // - contract.getTrait(...)
  // - contract.getXp(...)
  // - contract.getStaminaTimestamp(...)
  // - contract.getStaminaPoints(...)
  // - contract.getStaminaPointsFromTimestamp(...)
  // - contract.isStaminaFull(...)
  // - contract.getStaminaMaxWait(...)
  // - contract.getFightDataAndDrainStamina(...)
  // - contract.canRaid(...)
}

export function handleRoleAdminChanged(event: RoleAdminChanged): void {}

export function handleRoleGranted(event: RoleGranted): void {}

export function handleRoleRevoked(event: RoleRevoked): void {}

export function handleTransfer(event: Transfer): void {}
