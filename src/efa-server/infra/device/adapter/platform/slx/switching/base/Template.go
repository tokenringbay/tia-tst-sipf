package base

var macMoveDetectCreate = `
<config>
      <mac-address-table xmlns="urn:brocade.com:mgmt:brocade-mac-address-table">
         <mac-move>
            <mac-move-detect-enable/>
         </mac-move>
      </mac-address-table>
</config>
`

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
      <mac-address-table xmlns="urn:brocade.com:mgmt:brocade-mac-address-table">
         <learning-mode>conversational</learning-mode>
         <aging-time>
            <conversational-time-out>{{.mac_aging_conversational_timeout}}</conversational-time-out>
         </aging-time>
         <mac-move>
            <mac-move-limit>{{.mac_move_limit}}</mac-move-limit>
         </mac-move>
      </mac-address-table>
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

var macMoveDetectDelete = `
<config>
	<mac-address-table xmlns="urn:brocade.com:mgmt:brocade-mac-address-table">
         <mac-move>
		<mac-move-detect-enable operation="remove"/>
         </mac-move>
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
		<mac-move>
			<mac-move-limit operation="remove"/>
		</mac-move>
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
var mctClusterAddControlVlan = `
<config>
        <cluster xmlns="urn:brocade.com:mgmt:brocade-mct">
                <cluster-name>{{.cluster_name}}</cluster-name>
                <cluster-id>{{.cluster_id}}</cluster-id>
                <cluster-control-vlan>{{.cluster_control_vlan}}</cluster-control-vlan>
        </cluster>
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
		<df-load-balance></df-load-balance>
		<deploy></deploy>
	</cluster>
</config>
`

var configureSwitchHostName = `
<config>
        <system xmlns="urn:brocade.com:mgmt:brocade-ras">
                <switch-attributes>
                <host-name>{{.host_name}}</host-name>
                </switch-attributes>
        </system>
</config>
`

var unconfigureSwitchHostName = `
<config>
        <system xmlns="urn:brocade.com:mgmt:brocade-ras">
                <switch-attributes>
                <host-name operation="delete"/>
                </switch-attributes>
        </system>
</config>
`
