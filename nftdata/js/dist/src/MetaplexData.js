"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.MetaplexData = exports.SERUM_MAIN_NET = exports.MAIN_NET = void 0;
const web3_js_1 = require("@solana/web3.js");
const mpl_token_metadata_1 = require("@metaplex-foundation/mpl-token-metadata");
const mpl_core_1 = require("@metaplex-foundation/mpl-core");
const bn_js_1 = __importDefault(require("bn.js"));
exports.MAIN_NET = 'https://api.mainnet-beta.solana.com';
exports.SERUM_MAIN_NET = 'https://solana-api.projectserum.com';
class MetaplexData {
    static setRPC(addr) {
        MetaplexData.connection = new web3_js_1.Connection(addr);
    }
    static findNftByOwner(ownerAddr) {
        return __awaiter(this, void 0, void 0, function* () {
            let owner = new web3_js_1.PublicKey(ownerAddr);
            const accounts = yield mpl_core_1.TokenAccount.getTokenAccountsByOwner(MetaplexData.connection, owner);
            const metadataPdaLookups = accounts.reduce((memo, { data }) => {
                var _a;
                return ((_a = data.amount) === null || _a === void 0 ? void 0 : _a.eq(new bn_js_1.default(1))) ? [...memo, mpl_token_metadata_1.Metadata.getPDA(data.mint)] : memo;
            }, []);
            const metadataAddresses = yield Promise.all(metadataPdaLookups);
            const tokenInfo = yield mpl_core_1.Account.getInfos(MetaplexData.connection, metadataAddresses);
            return Array.from(tokenInfo.values()).map((m) => mpl_token_metadata_1.MetadataData.deserialize(m.data));
        });
    }
}
exports.MetaplexData = MetaplexData;
MetaplexData.connection = new web3_js_1.Connection(exports.SERUM_MAIN_NET);
//# sourceMappingURL=MetaplexData.js.map