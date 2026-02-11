import * as fcl from "@onflow/fcl";

export async function configureFCL(mainnetAddresses?: Record<string, string>) {
  fcl.config({
    "accessNode.api": "https://rest-mainnet.onflow.org",
    "flow.network": "mainnet",
  });

  // Register contract addresses for import resolution
  if (mainnetAddresses) {
    for (const [placeholder, real] of Object.entries(mainnetAddresses)) {
      fcl.config().put(placeholder, real);
    }
  }
}
