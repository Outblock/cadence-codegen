import * as fcl from "@onflow/fcl";
import { Buffer } from 'buffer';

/** Utility function to decode Base64 Cadence code */
const decodeCadence = (code: string): string => Buffer.from(code, 'base64').toString('utf8');

/** Generated from Cadence files */
/** Flow Signer interface for transaction signing */
export interface FlowSigner {
  address: string;
  keyIndex: number;
  sign(signableData: Uint8Array): Promise<Uint8Array>;
  authzFunc: (account: any) => Promise<any>;
}

export interface CompositeSignature {
  addr: string;
  keyId: number;
  signature: string;
}

export interface AuthorizationAccount extends Record<string, any> {
  tempId: string;
  addr: string;
  keyId: number;
  signingFunction: (signable: { message: string }) => Promise<CompositeSignature>;
}

export type AuthorizationFunction = (account: any) => Promise<AuthorizationAccount>;

/** Network addresses for contract imports */
export const addresses = {"mainnet":{"0xDomains":"0x233eb012d34b0070","0xEVM":"0xe467b9dd11fa00df","0xFind":"0x097bafa4e0b48eef","0xFlowEVMBridge":"0x1e4aa0b87d10b141","0xFlowFees":"0xf919ee77447b7497","0xFlowIDTableStaking":"0x8624b52f9ddcd04a","0xFlowStakingCollection":"0x8d0e87b65159ae63","0xFlowToken":"0x1654653399040a61","0xFlowns":"0x233eb012d34b0070","0xFungibleToken":"0xf233dcee88fe0abe","0xHybridCustody":"0xd8a7e05a7ac670c0","0xLockedTokens":"0x8d0e87b65159ae63","0xMetadataViews":"0x1d7e57aa55817448","0xNFTCatalog":"0x49a7cda3a1eecc29","0xNFTRetrieval":"0x49a7cda3a1eecc29","0xNonFungibleToken":"0x1d7e57aa55817448","0xStringUtils":"0xa340dc0a4ec828ab","0xTransactionGeneration":"0xe52522745adf5c34","0xViewResolver":"0x1d7e57aa55817448"},"testnet":{"0xDomains":"0xb05b2abb42335e88","0xEVM":"0x8c5303eaa26202d6","0xFind":"0xa16ab1d0abde3625","0xFlowEVMBridge":"0xdfc20aee650fcbdf","0xFlowFees":"0x912d5440f7e3769e","0xFlowToken":"0x7e60df042a9c0868","0xFlowns":"0xb05b2abb42335e88","0xFungibleToken":"0x9a0766d93b6608b7","0xHybridCustody":"0x294e44e1ec6993c6","0xMetadataViews":"0x631e88ae7f1d7c20","0xNFTCatalog":"0x324c34e1c517e4db","0xNFTRetrieval":"0x324c34e1c517e4db","0xNonFungibleToken":"0x631e88ae7f1d7c20","0xStringUtils":"0x31ad40c07a2a9788","0xTransactionGeneration":"0x830c495357676f8b","0xViewResolver":"0x631e88ae7f1d7c20"}};

/** Generated Cadence interface */
export interface StorageInfo {
    capacity: number;
    used: number;
    available: number;
}

/** Generated Cadence interface */
export interface DelegatorInfo {
    id: number;
    nodeID: string;
    tokensCommitted: string;
    tokensStaked: string;
    tokensUnstaking: string;
    tokensRewarded: string;
    tokensUnstaked: string;
    tokensRequestedToUnstake: string;
}

/** Generated Cadence interface */
export interface FlowIDTableStakingDelegatorInfo {
    id: number;
    nodeID: string;
    tokensCommitted: string;
    tokensStaked: string;
    tokensUnstaking: string;
    tokensRewarded: string;
    tokensUnstaked: string;
    tokensRequestedToUnstake: string;
}

type RequestInterceptor = (config: any) => any | Promise<any>;
type ResponseInterceptor = (response: any) => any | Promise<any>;

export class CadenceService {
  private requestInterceptors: RequestInterceptor[] = [];
  private responseInterceptors: ResponseInterceptor[] = [];

  constructor() {
  }

  useRequestInterceptor(interceptor: RequestInterceptor) {
    this.requestInterceptors.push(interceptor);
  }

  useResponseInterceptor(interceptor: ResponseInterceptor) {
    this.responseInterceptors.push(interceptor);
  }

  private async runRequestInterceptors(config: any) {
    let c = config;
    for (const interceptor of this.requestInterceptors) {
      c = await interceptor(c);
    }
    return c;
  }

  private async runResponseInterceptors(response: any) {
    let r = response;
    for (const interceptor of this.responseInterceptors) {
      r = await interceptor(r);
    }
    return r;
  }



  public async callContract(toEVMAddressHex: string, amount: string, data: number[], gasLimit: number) {
    const code = decodeCadence("aW1wb3J0IEZ1bmdpYmxlVG9rZW4gZnJvbSAweEZ1bmdpYmxlVG9rZW4KaW1wb3J0IEZsb3dUb2tlbiBmcm9tIDB4Rmxvd1Rva2VuCmltcG9ydCBFVk0gZnJvbSAweEVWTQoKLy8vIFRyYW5zZmVycyAkRkxPVyBmcm9tIHRoZSBzaWduZXIncyBhY2NvdW50IENhZGVuY2UgRmxvdyBiYWxhbmNlIHRvIHRoZSByZWNpcGllbnQncyBoZXgtZW5jb2RlZCBFVk0gYWRkcmVzcy4KLy8vIE5vdGUgdGhhdCBhIENPQSBtdXN0IGhhdmUgYSAkRkxPVyBiYWxhbmNlIGluIEVWTSBiZWZvcmUgdHJhbnNmZXJyaW5nIHZhbHVlIHRvIGFub3RoZXIgRVZNIGFkZHJlc3MuCi8vLwp0cmFuc2FjdGlvbih0b0VWTUFkZHJlc3NIZXg6IFN0cmluZywgYW1vdW50OiBVRml4NjQsIGRhdGE6IFtVSW50OF0sIGdhc0xpbWl0OiBVSW50NjQpIHsKCiAgICBsZXQgY29hOiBhdXRoKEVWTS5XaXRoZHJhdywgRVZNLkNhbGwpICZFVk0uQ2FkZW5jZU93bmVkQWNjb3VudAogICAgbGV0IHJlY2lwaWVudEVWTUFkZHJlc3M6IEVWTS5FVk1BZGRyZXNzCgogICAgcHJlcGFyZShzaWduZXI6IGF1dGgoQm9ycm93VmFsdWUsIFNhdmVWYWx1ZSkgJkFjY291bnQpIHsKICAgICAgICBpZiBzaWduZXIuc3RvcmFnZS50eXBlKGF0OiAvc3RvcmFnZS9ldm0pID09IG5pbCB7CiAgICAgICAgICAgIHNpZ25lci5zdG9yYWdlLnNhdmUoPC1FVk0uY3JlYXRlQ2FkZW5jZU93bmVkQWNjb3VudCgpLCB0bzogL3N0b3JhZ2UvZXZtKQogICAgICAgIH0KICAgICAgICBzZWxmLmNvYSA9IHNpZ25lci5zdG9yYWdlLmJvcnJvdzxhdXRoKEVWTS5XaXRoZHJhdywgRVZNLkNhbGwpICZFVk0uQ2FkZW5jZU93bmVkQWNjb3VudD4oZnJvbTogL3N0b3JhZ2UvZXZtKQogICAgICAgICAgICA/PyBwYW5pYygiQ291bGQgbm90IGJvcnJvdyByZWZlcmVuY2UgdG8gdGhlIHNpZ25lcidzIGJyaWRnZWQgYWNjb3VudCIpCgogICAgICAgIHNlbGYucmVjaXBpZW50RVZNQWRkcmVzcyA9IEVWTS5hZGRyZXNzRnJvbVN0cmluZyh0b0VWTUFkZHJlc3NIZXgpCiAgICB9CgogICAgZXhlY3V0ZSB7CiAgICAgICAgaWYgc2VsZi5yZWNpcGllbnRFVk1BZGRyZXNzLmJ5dGVzID09IHNlbGYuY29hLmFkZHJlc3MoKS5ieXRlcyB7CiAgICAgICAgICAgIHJldHVybgogICAgICAgIH0KICAgICAgICBsZXQgdmFsdWVCYWxhbmNlID0gRVZNLkJhbGFuY2UoYXR0b2Zsb3c6IDApCiAgICAgICAgdmFsdWVCYWxhbmNlLnNldEZMT1coZmxvdzogYW1vdW50KQogICAgICAgIGxldCB0eFJlc3VsdCA9IHNlbGYuY29hLmNhbGwoCiAgICAgICAgICAgIHRvOiBzZWxmLnJlY2lwaWVudEVWTUFkZHJlc3MsCiAgICAgICAgICAgIGRhdGE6IGRhdGEsCiAgICAgICAgICAgIGdhc0xpbWl0OiBnYXNMaW1pdCwKICAgICAgICAgICAgdmFsdWU6IHZhbHVlQmFsYW5jZQogICAgICAgICkKICAgICAgICBhc3NlcnQoCiAgICAgICAgICAgIHR4UmVzdWx0LnN0YXR1cyA9PSBFVk0uU3RhdHVzLmZhaWxlZCB8fCB0eFJlc3VsdC5zdGF0dXMgPT0gRVZNLlN0YXR1cy5zdWNjZXNzZnVsLAogICAgICAgICAgICBtZXNzYWdlOiAiZXZtX2Vycm9yPSIuY29uY2F0KHR4UmVzdWx0LmVycm9yTWVzc2FnZSkuY29uY2F0KCJcbiIpCiAgICAgICAgKQogICAgfQp9");
    let config = {
      cadence: code,
      name: "callContract",
      type: "transaction",
      args: (arg: any, t: any) => [
        arg(toEVMAddressHex, t.String),
        arg(amount, t.UFix64),
        arg(data, t.Array(t.UInt8)),
        arg(gasLimit, t.UInt64),
      ],
	  	limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let txId = await fcl.mutate(config);
    txId = await this.runResponseInterceptors(txId);
    return txId;
  }


  public async createCoa(amount: string) {
    const code = decodeCadence("aW1wb3J0IEZ1bmdpYmxlVG9rZW4gZnJvbSAweEZ1bmdpYmxlVG9rZW4KaW1wb3J0IEZsb3dUb2tlbiBmcm9tIDB4Rmxvd1Rva2VuCmltcG9ydCBFVk0gZnJvbSAweEVWTQoKCi8vLyBDcmVhdGVzIGEgQ09BIGFuZCBzYXZlcyBpdCBpbiB0aGUgc2lnbmVyJ3MgRmxvdyBhY2NvdW50ICYgcGFzc2luZyB0aGUgZ2l2ZW4gdmFsdWUgb2YgRmxvdyBpbnRvIEZsb3dFVk0KdHJhbnNhY3Rpb24oYW1vdW50OiBVRml4NjQpIHsKICAgIGxldCBzZW50VmF1bHQ6IEBGbG93VG9rZW4uVmF1bHQKICAgIGxldCBhdXRoOiBhdXRoKElzc3VlU3RvcmFnZUNhcGFiaWxpdHlDb250cm9sbGVyLCBJc3N1ZVN0b3JhZ2VDYXBhYmlsaXR5Q29udHJvbGxlciwgUHVibGlzaENhcGFiaWxpdHksIFNhdmVWYWx1ZSwgVW5wdWJsaXNoQ2FwYWJpbGl0eSkgJkFjY291bnQKCiAgICBwcmVwYXJlKHNpZ25lcjogYXV0aChCb3Jyb3dWYWx1ZSwgSXNzdWVTdG9yYWdlQ2FwYWJpbGl0eUNvbnRyb2xsZXIsIFB1Ymxpc2hDYXBhYmlsaXR5LCBTYXZlVmFsdWUsIFVucHVibGlzaENhcGFiaWxpdHkpICZBY2NvdW50KSB7CiAgICAgICAgbGV0IHZhdWx0UmVmID0gc2lnbmVyLnN0b3JhZ2UuYm9ycm93PGF1dGgoRnVuZ2libGVUb2tlbi5XaXRoZHJhdykgJkZsb3dUb2tlbi5WYXVsdD4oCiAgICAgICAgICAgICAgICBmcm9tOiAvc3RvcmFnZS9mbG93VG9rZW5WYXVsdAogICAgICAgICAgICApID8/IHBhbmljKCJDb3VsZCBub3QgYm9ycm93IHJlZmVyZW5jZSB0byB0aGUgb3duZXIncyBWYXVsdCEiKQoKICAgICAgICBzZWxmLnNlbnRWYXVsdCA8LSB2YXVsdFJlZi53aXRoZHJhdyhhbW91bnQ6IGFtb3VudCkgYXMhIEBGbG93VG9rZW4uVmF1bHQKICAgICAgICBzZWxmLmF1dGggPSBzaWduZXIKICAgIH0KCiAgICBleGVjdXRlIHsKICAgICAgICBsZXQgY29hIDwtIEVWTS5jcmVhdGVDYWRlbmNlT3duZWRBY2NvdW50KCkKICAgICAgICBjb2EuZGVwb3NpdChmcm9tOiA8LXNlbGYuc2VudFZhdWx0KQoKICAgICAgICBsb2coY29hLmJhbGFuY2UoKS5pbkZMT1coKSkKICAgICAgICBsZXQgc3RvcmFnZVBhdGggPSBTdG9yYWdlUGF0aChpZGVudGlmaWVyOiAiZXZtIikhCiAgICAgICAgbGV0IHB1YmxpY1BhdGggPSBQdWJsaWNQYXRoKGlkZW50aWZpZXI6ICJldm0iKSEKICAgICAgICBzZWxmLmF1dGguc3RvcmFnZS5zYXZlPEBFVk0uQ2FkZW5jZU93bmVkQWNjb3VudD4oPC1jb2EsIHRvOiBzdG9yYWdlUGF0aCkKICAgICAgICBsZXQgYWRkcmVzc2FibGVDYXAgPSBzZWxmLmF1dGguY2FwYWJpbGl0aWVzLnN0b3JhZ2UuaXNzdWU8JkVWTS5DYWRlbmNlT3duZWRBY2NvdW50PihzdG9yYWdlUGF0aCkKICAgICAgICBzZWxmLmF1dGguY2FwYWJpbGl0aWVzLnVucHVibGlzaChwdWJsaWNQYXRoKQogICAgICAgIHNlbGYuYXV0aC5jYXBhYmlsaXRpZXMucHVibGlzaChhZGRyZXNzYWJsZUNhcCwgYXQ6IHB1YmxpY1BhdGgpCiAgICB9Cn0=");
    let config = {
      cadence: code,
      name: "createCoa",
      type: "transaction",
      args: (arg: any, t: any) => [
        arg(amount, t.UFix64),
      ],
	  	limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let txId = await fcl.mutate(config);
    txId = await this.runResponseInterceptors(txId);
    return txId;
  }

  public async getAddr(flowAddress: string): Promise<string | undefined> {
    const code = decodeCadence("aW1wb3J0IEVWTSBmcm9tIDB4RVZNCgphY2Nlc3MoYWxsKSBmdW4gbWFpbihmbG93QWRkcmVzczogQWRkcmVzcyk6IFN0cmluZz8gewogICAgaWYgbGV0IGFkZHJlc3M6IEVWTS5FVk1BZGRyZXNzID0gZ2V0QXV0aEFjY291bnQ8YXV0aChCb3Jyb3dWYWx1ZSkgJkFjY291bnQ+KGZsb3dBZGRyZXNzKQogICAgICAgIC5zdG9yYWdlLmJvcnJvdzwmRVZNLkNhZGVuY2VPd25lZEFjY291bnQ+KGZyb206IC9zdG9yYWdlL2V2bSk/LmFkZHJlc3MoKSB7CiAgICAgICAgbGV0IGJ5dGVzOiBbVUludDhdID0gW10KICAgICAgICBmb3IgYnl0ZSBpbiBhZGRyZXNzLmJ5dGVzIHsKICAgICAgICAgICAgYnl0ZXMuYXBwZW5kKGJ5dGUpCiAgICAgICAgfQogICAgICAgIHJldHVybiBTdHJpbmcuZW5jb2RlSGV4KGJ5dGVzKQogICAgfQogICAgcmV0dXJuIG5pbAp9");
    let config = {
      cadence: code,
      name: "getAddr",
      type: "script",
      args: (arg: any, t: any) => [
        arg(flowAddress, t.Address),
      ],
	  	limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let result = await fcl.query(config);
    result = await this.runResponseInterceptors(result);
    return result;
  }

  public async getChildAccountMeta(parent: string): Promise<Record<string, any>> {
    const code = decodeCadence("aW1wb3J0IEh5YnJpZEN1c3RvZHkgZnJvbSAweEh5YnJpZEN1c3RvZHkKaW1wb3J0IE1ldGFkYXRhVmlld3MgZnJvbSAweE1ldGFkYXRhVmlld3MKCmFjY2VzcyhhbGwpIGZ1biBtYWluKHBhcmVudDogQWRkcmVzcyk6IHtBZGRyZXNzOiBBbnlTdHJ1Y3R9IHsKICAgIGxldCBhY2N0ID0gZ2V0QXV0aEFjY291bnQ8YXV0aChTdG9yYWdlKSAmQWNjb3VudD4ocGFyZW50KQogICAgbGV0IG0gPSBhY2N0LnN0b3JhZ2UuYm9ycm93PCZIeWJyaWRDdXN0b2R5Lk1hbmFnZXI+KGZyb206IEh5YnJpZEN1c3RvZHkuTWFuYWdlclN0b3JhZ2VQYXRoKQoKICAgIGlmIG0gPT0gbmlsIHsKICAgICAgICByZXR1cm4ge30KICAgIH0gZWxzZSB7CiAgICAgICAgdmFyIGRhdGE6IHtBZGRyZXNzOiBBbnlTdHJ1Y3R9ID0ge30KICAgICAgICBmb3IgYWRkcmVzcyBpbiBtPy5nZXRDaGlsZEFkZHJlc3NlcygpISB7CiAgICAgICAgICAgIGxldCBjID0gbT8uZ2V0Q2hpbGRBY2NvdW50RGlzcGxheShhZGRyZXNzOiBhZGRyZXNzKSAKICAgICAgICAgICAgZGF0YS5pbnNlcnQoa2V5OiBhZGRyZXNzLCBjKQogICAgICAgIH0KICAgICAgICByZXR1cm4gZGF0YQogICAgfQp9Cg==");
    let config = {
      cadence: code,
      name: "getChildAccountMeta",
      type: "script",
      args: (arg: any, t: any) => [
        arg(parent, t.Address),
      ],
	  	limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let result = await fcl.query(config);
    result = await this.runResponseInterceptors(result);
    return result;
  }


  public async getChildAddresses(parent: string): Promise<string[]> {
    const code = decodeCadence("aW1wb3J0IEh5YnJpZEN1c3RvZHkgZnJvbSAweEh5YnJpZEN1c3RvZHkKCmFjY2VzcyhhbGwpIGZ1biBtYWluKHBhcmVudDogQWRkcmVzcyk6IFtBZGRyZXNzXSB7CiAgICBsZXQgYWNjdCA9IGdldEF1dGhBY2NvdW50PGF1dGgoU3RvcmFnZSkgJkFjY291bnQ+KHBhcmVudCkKICAgIGlmIGxldCBtYW5hZ2VyID0gYWNjdC5zdG9yYWdlLmJvcnJvdzwmSHlicmlkQ3VzdG9keS5NYW5hZ2VyPihmcm9tOiBIeWJyaWRDdXN0b2R5Lk1hbmFnZXJTdG9yYWdlUGF0aCkgewogICAgICAgIHJldHVybiAgbWFuYWdlci5nZXRDaGlsZEFkZHJlc3NlcygpCiAgICB9CiAgICByZXR1cm4gW10KfQo=");
    let config = {
      cadence: code,
      name: "getChildAddresses",
      type: "script",
      args: (arg: any, t: any) => [
        arg(parent, t.Address),
      ],
	  	limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let result = await fcl.query(config);
    result = await this.runResponseInterceptors(result);
    return result;
  }

  public async getDelegator(address: string): Promise<DelegatorInfo[] | undefined> {
    const code = decodeCadence("aW1wb3J0IEZsb3dTdGFraW5nQ29sbGVjdGlvbiBmcm9tIDB4Rmxvd1N0YWtpbmdDb2xsZWN0aW9uCmltcG9ydCBGbG93SURUYWJsZVN0YWtpbmcgZnJvbSAweEZsb3dJRFRhYmxlU3Rha2luZwppbXBvcnQgTG9ja2VkVG9rZW5zIGZyb20gMHhMb2NrZWRUb2tlbnMKCiBhY2Nlc3MoYWxsKSBzdHJ1Y3QgRGVsZWdhdG9ySW5mbyB7CiAgICBhY2Nlc3MoYWxsKSBsZXQgaWQ6IFVJbnQzMgogICAgYWNjZXNzKGFsbCkgbGV0IG5vZGVJRDogU3RyaW5nCiAgICBhY2Nlc3MoYWxsKSBsZXQgdG9rZW5zQ29tbWl0dGVkOiBVRml4NjQKICAgIGFjY2VzcyhhbGwpIGxldCB0b2tlbnNTdGFrZWQ6IFVGaXg2NAogICAgYWNjZXNzKGFsbCkgbGV0IHRva2Vuc1Vuc3Rha2luZzogVUZpeDY0CiAgICBhY2Nlc3MoYWxsKSBsZXQgdG9rZW5zUmV3YXJkZWQ6IFVGaXg2NAogICAgYWNjZXNzKGFsbCkgbGV0IHRva2Vuc1Vuc3Rha2VkOiBVRml4NjQKICAgIGFjY2VzcyhhbGwpIGxldCB0b2tlbnNSZXF1ZXN0ZWRUb1Vuc3Rha2U6IFVGaXg2NAp9CgphY2Nlc3MoYWxsKSBmdW4gbWFpbihhZGRyZXNzOiBBZGRyZXNzKTogW0RlbGVnYXRvckluZm9dPyB7CiAgICB2YXIgcmVzOiBbRGVsZWdhdG9ySW5mb10/ID0gbmlsCgogICAgbGV0IGluaXRlZCA9IEZsb3dTdGFraW5nQ29sbGVjdGlvbi5kb2VzQWNjb3VudEhhdmVTdGFraW5nQ29sbGVjdGlvbihhZGRyZXNzOiBhZGRyZXNzKQoKICAgIGlmIGluaXRlZCB7CiAgICAgICAgbGV0IHJlc3VsdCA9IEZsb3dTdGFraW5nQ29sbGVjdGlvbi5nZXRBbGxEZWxlZ2F0b3JJbmZvKGFkZHJlc3M6IGFkZHJlc3MpCiAgICAgICAgZm9yIGluZm8gaW4gcmVzdWx0IHsKICAgICAgICAgICAgcmVzLmFwcGVuZChEZWxlZ2F0b3JJbmZvKGlkOiBpbmZvLmlkLCBub2RlSUQ6IGluZm8ubm9kZUlELCB0b2tlbnNDb21taXR0ZWQ6IGluZm8udG9rZW5zQ29tbWl0dGVkLCB0b2tlbnNTdGFrZWQ6IGluZm8udG9rZW5zU3Rha2VkLCB0b2tlbnNVbnN0YWtpbmc6IGluZm8udG9rZW5zVW5zdGFraW5nLCB0b2tlbnNSZXdhcmRlZDogaW5mby50b2tlbnNSZXdhcmRlZCwgdG9rZW5zVW5zdGFrZWQ6IGluZm8udG9rZW5zVW5zdGFrZWQsIHRva2Vuc1JlcXVlc3RlZFRvVW5zdGFrZTogaW5mby50b2tlbnNSZXF1ZXN0ZWRUb1Vuc3Rha2UpKQogICAgICAgIH0KICAgIH0KICAgIHJldHVybiByZXMKfQo=");
    let config = {
      cadence: code,
      name: "getDelegator",
      type: "script",
      args: (arg: any, t: any) => [
        arg(address, t.Address),
      ],
	  	limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let result = await fcl.query(config);
    result = await this.runResponseInterceptors(result);
    return result;
  }

  public async getDelegatorInfo(address: string): Promise<FlowIDTableStakingDelegatorInfo[] | undefined> {
    const code = decodeCadence("aW1wb3J0IEZsb3dTdGFraW5nQ29sbGVjdGlvbiBmcm9tIDB4Rmxvd1N0YWtpbmdDb2xsZWN0aW9uCmltcG9ydCBGbG93SURUYWJsZVN0YWtpbmcgZnJvbSAweEZsb3dJRFRhYmxlU3Rha2luZwppbXBvcnQgTG9ja2VkVG9rZW5zIGZyb20gMHhMb2NrZWRUb2tlbnMKICAgICAgICAKYWNjZXNzKGFsbCkgZnVuIG1haW4oYWRkcmVzczogQWRkcmVzcyk6IFtGbG93SURUYWJsZVN0YWtpbmcuRGVsZWdhdG9ySW5mb10/IHsKICAgIHZhciByZXM6IFtGbG93SURUYWJsZVN0YWtpbmcuRGVsZWdhdG9ySW5mb10/ID0gbmlsCgogICAgbGV0IGluaXRlZCA9IEZsb3dTdGFraW5nQ29sbGVjdGlvbi5kb2VzQWNjb3VudEhhdmVTdGFraW5nQ29sbGVjdGlvbihhZGRyZXNzOiBhZGRyZXNzKQoKICAgIGlmIGluaXRlZCB7CiAgICAgICAgcmVzID0gRmxvd1N0YWtpbmdDb2xsZWN0aW9uLmdldEFsbERlbGVnYXRvckluZm8oYWRkcmVzczogYWRkcmVzcykKICAgIH0KICAgIHJldHVybiByZXMKfQo=");
    let config = {
      cadence: code,
      name: "getDelegatorInfo",
      type: "script",
      args: (arg: any, t: any) => [
        arg(address, t.Address),
      ],
	  	limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let result = await fcl.query(config);
    result = await this.runResponseInterceptors(result);
    return result;
  }

  public async getTokenBalanceStorage(address: string): Promise<Record<string, string>> {
    const code = decodeCadence("aW1wb3J0IEZ1bmdpYmxlVG9rZW4gZnJvbSAweEZ1bmdpYmxlVG9rZW4KCi8vLyBRdWVyaWVzIGZvciBGVC5WYXVsdCBiYWxhbmNlIG9mIGFsbCBGVC5WYXVsdHMgaW4gdGhlIHNwZWNpZmllZCBhY2NvdW50LgovLy8KYWNjZXNzKGFsbCkgZnVuIG1haW4oYWRkcmVzczogQWRkcmVzcyk6IHtTdHJpbmc6IFVGaXg2NH0gewogICAgLy8gR2V0IHRoZSBhY2NvdW50CiAgICBsZXQgYWNjb3VudCA9IGdldEF1dGhBY2NvdW50PGF1dGgoQm9ycm93VmFsdWUpICZBY2NvdW50PihhZGRyZXNzKQogICAgLy8gSW5pdCBmb3IgcmV0dXJuIHZhbHVlCiAgICBsZXQgYmFsYW5jZXM6IHtTdHJpbmc6IFVGaXg2NH0gPSB7fQogICAgLy8gVHJhY2sgc2VlbiBUeXBlcyBpbiBhcnJheQogICAgbGV0IHNlZW46IFtTdHJpbmddID0gW10KICAgIC8vIEFzc2lnbiB0aGUgdHlwZSB3ZSdsbCBuZWVkCiAgICBsZXQgdmF1bHRUeXBlOiBUeXBlID0gVHlwZTxAe0Z1bmdpYmxlVG9rZW4uVmF1bHR9PigpCiAgICAvLyBJdGVyYXRlIG92ZXIgYWxsIHN0b3JlZCBpdGVtcyAmIGdldCB0aGUgcGF0aCBpZiB0aGUgdHlwZSBpcyB3aGF0IHdlJ3JlIGxvb2tpbmcgZm9yCiAgICBhY2NvdW50LnN0b3JhZ2UuZm9yRWFjaFN0b3JlZChmdW4gKHBhdGg6IFN0b3JhZ2VQYXRoLCB0eXBlOiBUeXBlKTogQm9vbCB7CiAgICAgICAgaWYgIXR5cGUuaXNSZWNvdmVyZWQgJiYgKHR5cGUuaXNJbnN0YW5jZSh2YXVsdFR5cGUpIHx8IHR5cGUuaXNTdWJ0eXBlKG9mOiB2YXVsdFR5cGUpKSB7CiAgICAgICAgICAgIC8vIEdldCBhIHJlZmVyZW5jZSB0byB0aGUgcmVzb3VyY2UgJiBpdHMgYmFsYW5jZQogICAgICAgICAgICBsZXQgdmF1bHRSZWYgPSBhY2NvdW50LnN0b3JhZ2UuYm9ycm93PCZ7RnVuZ2libGVUb2tlbi5CYWxhbmNlfT4oZnJvbTogcGF0aCkhCiAgICAgICAgICAgIC8vIEluc2VydCBhIG5ldyB2YWx1ZXMgaWYgaXQncyB0aGUgZmlyc3QgdGltZSB3ZSd2ZSBzZWVuIHRoZSB0eXBlCiAgICAgICAgICAgIGlmICFzZWVuLmNvbnRhaW5zKHR5cGUuaWRlbnRpZmllcikgewogICAgICAgICAgICAgICAgYmFsYW5jZXMuaW5zZXJ0KGtleTogdHlwZS5pZGVudGlmaWVyLCB2YXVsdFJlZi5iYWxhbmNlKQogICAgICAgICAgICB9IGVsc2UgewogICAgICAgICAgICAgICAgLy8gT3RoZXJ3aXNlIGp1c3QgdXBkYXRlIHRoZSBiYWxhbmNlIG9mIHRoZSB2YXVsdCAodW5saWtlbHkgd2UnbGwgc2VlIHRoZSBzYW1lIHR5cGUgdHdpY2UgaW4KICAgICAgICAgICAgICAgIC8vIHRoZSBzYW1lIGFjY291bnQsIGJ1dCB3ZSB3YW50IHRvIGNvdmVyIHRoZSBjYXNlKQogICAgICAgICAgICAgICAgYmFsYW5jZXNbdHlwZS5pZGVudGlmaWVyXSA9IGJhbGFuY2VzW3R5cGUuaWRlbnRpZmllcl0hICsgdmF1bHRSZWYuYmFsYW5jZQogICAgICAgICAgICB9CiAgICAgICAgfQogICAgICAgIHJldHVybiB0cnVlCiAgICB9KQoKICAgIC8vIEFkZCBhdmFpbGFibGUgRmxvdyBUb2tlbiBCYWxhbmNlCiAgICBiYWxhbmNlcy5pbnNlcnQoa2V5OiAiYXZhaWxhYmxlRmxvd1Rva2VuIiwgYWNjb3VudC5hdmFpbGFibGVCYWxhbmNlKQoKICAgIHJldHVybiBiYWxhbmNlcwp9");
    let config = {
      cadence: code,
      name: "getTokenBalanceStorage",
      type: "script",
      args: (arg: any, t: any) => [
        arg(address, t.Address),
      ],
	  	limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let result = await fcl.query(config);
    result = await this.runResponseInterceptors(result);
    return result;
  }

  public async accountStorage(addr: string): Promise<StorageInfo> {
    const code = decodeCadence("YWNjZXNzKGFsbCkgCnN0cnVjdCBTdG9yYWdlSW5mbyB7CiAgICBhY2Nlc3MoYWxsKSBsZXQgY2FwYWNpdHk6IFVJbnQ2NAogICAgYWNjZXNzKGFsbCkgbGV0IHVzZWQ6IFVJbnQ2NAogICAgYWNjZXNzKGFsbCkgbGV0IGF2YWlsYWJsZTogVUludDY0CgogICAgaW5pdChjYXBhY2l0eTogVUludDY0LCB1c2VkOiBVSW50NjQsIGF2YWlsYWJsZTogVUludDY0KSB7CiAgICAgICAgc2VsZi5jYXBhY2l0eSA9IGNhcGFjaXR5CiAgICAgICAgc2VsZi51c2VkID0gdXNlZAogICAgICAgIHNlbGYuYXZhaWxhYmxlID0gYXZhaWxhYmxlCiAgICB9Cn0KCmFjY2VzcyhhbGwpIGZ1biBtYWluKGFkZHI6IEFkZHJlc3MpOiBTdG9yYWdlSW5mbyB7CiAgICBsZXQgYWNjdDogJkFjY291bnQgPSBnZXRBY2NvdW50KGFkZHIpCiAgICByZXR1cm4gU3RvcmFnZUluZm8oY2FwYWNpdHk6IGFjY3Quc3RvcmFnZS5jYXBhY2l0eSwKICAgICAgICAgICAgICAgICAgICAgIHVzZWQ6IGFjY3Quc3RvcmFnZS51c2VkLAogICAgICAgICAgICAgICAgICAgICAgYXZhaWxhYmxlOiBhY2N0LnN0b3JhZ2UuY2FwYWNpdHkgLSBhY2N0LnN0b3JhZ2UudXNlZCkKfSA=");
    let config = {
      cadence: code,
      name: "accountStorage",
      type: "script",
      args: (arg: any, t: any) => [
        arg(addr, t.Address),
      ],
	  	limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let result = await fcl.query(config);
    result = await this.runResponseInterceptors(result);
    return result;
  }}
