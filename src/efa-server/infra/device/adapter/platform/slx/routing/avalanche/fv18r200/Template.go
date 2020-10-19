package fv18r200

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
