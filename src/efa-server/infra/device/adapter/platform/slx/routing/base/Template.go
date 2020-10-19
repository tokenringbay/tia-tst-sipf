package base

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
		<learning-mode operation="remove"/>
		<aging-time>
			<legacy-time-out operation="remove"/>
		</aging-time>
	</mac-address-table>
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

var routerMPLSCreate = `
<config>
	<mpls-config xmlns="urn:brocade.com:mgmt:brocade-mpls">
         <router>
            <mpls></mpls>
         </router>
      </mpls-config>
</config>
`
var routerMPLSDelete = `
<config>
	<mpls-config xmlns="urn:brocade.com:mgmt:brocade-mpls" >
         <router operation="remove">
            <mpls></mpls>
         </router>
      </mpls-config>
</config>
`

var intPortChannelAddSwitchPortBasic = `
<config>
	<interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <port-channel>
            <name>{{.name}}</name>
			<switchport-basic><basic/></switchport-basic>
         </port-channel>
      </interface>
</config>
`

var intPortChannelSwitchPortVlanMode = `
<config>
	<interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <port-channel>
            <name>{{.name}}</name>
			<switchport>
				<mode>
                	<vlan-mode>trunk-no-default-native</vlan-mode>
            	</mode>
			</switchport>
         </port-channel>
      </interface>
</config>
`

var intPortChannelSwitchPortAddVlan = `
<config>
	<interface xmlns="urn:brocade.com:mgmt:brocade-interface">
		<port-channel>
            <name>{{.name}}</name>
			<switchport>
            	<trunk>
            		<allowed>
            			<vlan>
            				<add>{{.vlan}}</add>
						</vlan>
					</allowed>
				</trunk>
			</switchport>
		</port-channel>
	</interface>
</config>
`

var configureSwitchHostName = `
<config>
        <system-ras xmlns="urn:brocade.com:mgmt:brocade-ras">
                <switch-attributes>
                <host-name>{{.host_name}}</host-name>
                </switch-attributes>
        </system-ras>
</config>
`

var unconfigureSwitchHostName = `
<config>
        <system-ras xmlns="urn:brocade.com:mgmt:brocade-ras">
                <switch-attributes>
                <host-name operation="delete"/>
                </switch-attributes>
        </system-ras>
</config>
`
