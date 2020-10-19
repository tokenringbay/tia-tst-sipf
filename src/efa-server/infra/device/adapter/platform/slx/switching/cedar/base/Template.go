package base

var ipAnycastGatewayCreate = `
<config>  
	<routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <ip>
            <static-ag-ip-config xmlns="urn:brocade.com:mgmt:brocade-vrrp">
               <anycast-gateway-mac>
                  <ip-anycast-gateway-mac>{{.ipv4_anycast_gateway_mac}}</ip-anycast-gateway-mac>
               </anycast-gateway-mac>
            </static-ag-ip-config>
         </ip>
      </routing-system>
</config>
`

var ipAnycastGatewayDelete = `
<config>  
	<routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <ip>
            <static-ag-ip-config xmlns="urn:brocade.com:mgmt:brocade-vrrp">
               <anycast-gateway-mac>
                  <ip-anycast-gateway-mac operation="remove"></ip-anycast-gateway-mac>
               </anycast-gateway-mac>
            </static-ag-ip-config>
         </ip>
      </routing-system>
</config>
`
