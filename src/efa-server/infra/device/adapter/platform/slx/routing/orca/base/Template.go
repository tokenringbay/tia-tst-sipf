package base

var conversationPropertyCreate = `
<config>
      <host-table xmlns="urn:brocade.com:mgmt:brocade-arp">
         <aging-mode>
            <conversational></conversational>
         </aging-mode>
         <aging-time>
            <conversational-timeout>{{.arp_aging_timeout}}</conversational-timeout>
         </aging-time>
      </host-table>
</config>
`

var configureLegacyMacTimeout = `
<config>
      <mac-address-table xmlns="urn:brocade.com:mgmt:brocade-mac-address-table">
         <aging-time>
            <legacy-time-out>{{.mac_aging_timeout}}</legacy-time-out>
         </aging-time>
      </mac-address-table>
</config>
`

var macConfigDelete = `
<config>
	<mac-address-table xmlns="urn:brocade.com:mgmt:brocade-mac-address-table">
		<aging-time>
			<legacy-time-out operation="remove"/>
		</aging-time>
	</mac-address-table>
</config>
`

var conversationMacDelete = `
<config>
	<mac-address-table xmlns="urn:brocade.com:mgmt:brocade-mac-address-table">
		<aging-time>
			<conversational-time-out operation="remove"/>
		</aging-time>
	</mac-address-table>
</config>
`

var conversationArpDelete = `
<config>
    <host-table xmlns="urn:brocade.com:mgmt:brocade-arp">
        <aging-time>
            <conversational-timeout operation="remove"/>
        </aging-time>
        <aging-mode operation="remove"/>
    </host-table>
</config>
`

var mctClusterAddPeer = `
<config>
	<cluster xmlns="urn:brocade.com:mgmt:brocade-mct">
		<cluster-name>{{.cluster_name}}</cluster-name>
		<cluster-id>{{.cluster_id}}</cluster-id>
		<peer-interface>
			<peer-if-type>{{.peer_if_type}}</peer-if-type>
			<peer-if-name>{{.peer_if_name}}</peer-if-name>
		</peer-interface>
		<peer>
			<peer-ip>{{.peer_ip}}</peer-ip>
              <source>
               <source_ip>{{.source_ip}}</source_ip>
              </source>
		</peer>
        
		<client-isolation>
            <loose></loose>
		</client-isolation>	
		<deploy></deploy>
	</cluster>
</config>
`

var mctClusterDelete = `
<config>
  <cluster xmlns="urn:brocade.com:mgmt:brocade-mct" operation="remove">
         <cluster-name>{{.cluster_name}}</cluster-name>
         <cluster-id>{{.cluster_id}}</cluster-id>
   </cluster>
</config>
`

var intVeCreate = `
  <config>
     <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
            <ve>
               <name>{{.name}}</name>
               <ip xmlns="urn:brocade.com:mgmt:brocade-ip-config">
                  <ip-config>
                     <address>
                        <address>{{.ip_address}}</address>
                     </address>
                  </ip-config>
               </ip>
               <bfd>
					<interval>
					  <min-tx>{{.bfd_min_tx}}</min-tx>
					  <min-rx>{{.bfd_min_rx}}</min-rx>
					  <multiplier>{{.bfd_multiplier}}</multiplier>
					</interval> 	
                </bfd>	
            </ve>
         </interface>
    </routing-system>
  </config>
`

var intVeActivate = `
<config>  
	  <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
          <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
           <ve>
               <name>{{.name}}</name>
               <shutdown xmlns="urn:brocade.com:mgmt:brocade-ip-config" operation="remove"></shutdown>
            </ve>
         </interface>
      </routing-system>
</config>
`
var configureSystemWideIPMtu = `
<config>
	<global-mtu-conf xmlns="urn:brocade.com:mgmt:brocade-interface">
		<ip xmlns="urn:brocade.com:mgmt:brocade-ip-config">
			<global-ip-mtu>{{.ip_mtu_value}}</global-ip-mtu>
		</ip>
	</global-mtu-conf>
</config>
`

var unconfigureSystemWideIPMtu = `
<config>
	<global-mtu-conf xmlns="urn:brocade.com:mgmt:brocade-interface">
		<ip xmlns="urn:brocade.com:mgmt:brocade-ip-config">
			<global-ip-mtu operation="remove"/>
		</ip>
	</global-mtu-conf>
</config>
`
