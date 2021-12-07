import { 
    AccountInfo, 
    Connection, 
    PublicKey } from '@solana/web3.js';

import {
    MetadataData,
    Metadata,
} from '@metaplex-foundation/mpl-token-metadata';

import {
    config,
    Borsh,
    Account,
    ERROR_INVALID_ACCOUNT_DATA,
    ERROR_INVALID_OWNER,
    AnyPublicKey,
    StringPublicKey,
    TokenAccount,
  } from '@metaplex-foundation/mpl-core';

import BN from 'bn.js';

export  const MAIN_NET = 'https://api.mainnet-beta.solana.com'
export  const SERUM_MAIN_NET = 'https://solana-api.projectserum.com'

export class MetaplexData  {

  static connection = new Connection(MAIN_NET)

  static setRPC(addr: string) {
    MetaplexData.connection = new Connection(addr)
  }

  static async findDataByOwner(
    ownerAddr: string,
  ): Promise<MetadataData[]> {
    let owner =  new PublicKey(ownerAddr);
    const accounts = await TokenAccount.getTokenAccountsByOwner(MetaplexData.connection, owner);

    const metadataPdaLookups = accounts.reduce((memo, { data }) => {
      // Only include tokens where amount equal to 1.
      // Note: This is not the same as mint supply.
      // NFTs by definition have supply of 1, but an account balance > 1 implies a mint supply > 1.
      return data.amount?.eq(new BN(1)) ? [...memo, Metadata.getPDA(data.mint)] : memo;
    }, []);

    const metadataAddresses = await Promise.all(metadataPdaLookups);
    const tokenInfo = await Account.getInfos(MetaplexData.connection, metadataAddresses);
    return Array.from(tokenInfo.values()).map((m) => MetadataData.deserialize(m.data));
  }
}
  