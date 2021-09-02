import { Address, BigInt } from "@graphprotocol/graph-ts"
import {
  Contract,
  Approval,
  ApprovalForAll,
  Transfer
} from "../generated/Contract/Contract"
import { TokenInfo,UserInfo,Dashboard } from "../generated/schema"

export function handleApproval(event: Approval): void {

}

export function handleApprovalForAll(event: ApprovalForAll): void {}

export function handleTransfer(event: Transfer): void {
  let token= TokenInfo.load(event.params.tokenId.toHex())
  if (token == null) {
    token = new TokenInfo(event.params.tokenId.toHex())

  }
  token.index = event.params.tokenId 
  token.owner = event.params.to


  let userFrom = UserInfo.load(event.params.from.toHex())
  if (userFrom == null) {
    userFrom = new UserInfo(event.params.from.toHex())
    userFrom.tokens = new Array()
  }

  let toFrom = UserInfo.load(event.params.to.toHex())
  if (toFrom == null) {
    toFrom = new UserInfo(event.params.to.toHex())
    toFrom.tokens = new Array()
  }

  userFrom.tokens = userFrom.tokens.filter(item => item !== event.params.tokenId);
  toFrom.tokens.push(event.params.tokenId)

  let dashboard = Dashboard.load(event.address.toHex())
  if (dashboard == null) {
    dashboard = new Dashboard(event.address.toHex())
    dashboard.burnCount = BigInt.fromI32(0)
    dashboard.transferCount = BigInt.fromI32(0)
    dashboard.mintCount = BigInt.fromI32(0)
  } 
  let zeroAddr = "0x0000000000000000000000000000000000000000"
  let one = BigInt.fromI32(1)
  if (event.params.from.toHex() == zeroAddr) {
    dashboard.mintCount =dashboard.mintCount.plus(one)
  } else  if (event.params.to.toHex() == zeroAddr ) {
    dashboard.burnCount =  dashboard.burnCount.plus(one)
  } else {
    dashboard.transferCount = dashboard.transferCount.plus(one)
  }
  token.save()
  userFrom.save()
  toFrom.save()
  dashboard.save()

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
  // - contract.balanceOf(...)
  // - contract.getApproved(...)
  // - contract.isApprovedForAll(...)
  // - contract.ownerOf(...)
  // - contract.supportsInterface(...)
}
