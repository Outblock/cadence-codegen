import FlowDomainUtils from 0xFlowDomainUtils

access(all) fun main(address: Address): {String: String} {
  return FlowDomainUtils.getDefaultDomainsOfAddress(address)
}