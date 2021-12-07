import { Connection } from '@solana/web3.js';
import { MetadataData } from '@metaplex-foundation/mpl-token-metadata';
export declare const MAIN_NET = "https://api.mainnet-beta.solana.com";
export declare const SERUM_MAIN_NET = "https://solana-api.projectserum.com";
export declare class MetaplexData {
    static connection: Connection;
    static setRPC(addr: string): void;
    static findDataByOwner(ownerAddr: string): Promise<MetadataData[]>;
}
