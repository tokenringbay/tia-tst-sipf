package base

var routerBgpMctNeighborCreate = `
<config>
    <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
    <router>
        <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
			
		<router-bgp-attributes>
		{{if  ne .neighborAddress  ""}}
			<neighbor>       
			<neighbor-ips>
				<neighbor-addr>
					<router-bgp-neighbor-address>{{.neighborAddress}}</router-bgp-neighbor-address>
					<remote-as>{{.remoteAs}}</remote-as>
                    <update-source>
                       <loopback>{{.loopbackNumber}}</loopback>
					</update-source>
					{{if eq .bfdEnabled "true"}}
					<bfd>
						<bfd-enable></bfd-enable>
					</bfd>
					{{end}}
				</neighbor-addr>
			</neighbor-ips>
		  </neighbor>
		{{end}}
		</router-bgp-attributes>
	
        </router-bgp>
    </router>
    </routing-system>
</config>  
`

var routerBgpMctL2EvpnNeighborCreate = `
<config>
    <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
    <router>
        <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
			
		<address-family>
		  <l2vpn>
			 <evpn>
				{{if  ne .neighborAddress  ""}}
				<neighbor>
				   <evpn-neighbor-ipv4>
					  <evpn-neighbor-ipv4-address>{{.neighborAddress}}</evpn-neighbor-ipv4-address>
					  <activate></activate>
					  {{if ne .encapType ""}}
					  <encapsulation>{{.encapType}}</encapsulation>
					  {{end}}
				   </evpn-neighbor-ipv4>
				</neighbor>
                 <graceful-restart>
			          <graceful-restart-status>
                      </graceful-restart-status>
                 </graceful-restart>
				{{end}}
			 </evpn>
		  </l2vpn>
		</address-family>
			   
        </router-bgp>
    </router>
    </routing-system>
</config>  
`

var routerBgpNeighborDeactivateInIpv4UnicastAF = `
<config>
    <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
    <router>
        <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
			
		<address-family>
		<ipv4>
			<ipv4-unicast>
				<default-vrf>
				   <default-vrf-selected></default-vrf-selected>
				   {{if  ne .peer_group_name  ""}}
				   <neighbor>
					  <af-ipv4-neighbor-peergroup-holder>
						 <af-ipv4-neighbor-peergroup>
							<af-ipv4-neighbor-peergroup-name>{{.peer_group_name}}</af-ipv4-neighbor-peergroup-name>
							<activate operation="remove"/>
						 </af-ipv4-neighbor-peergroup>
					  </af-ipv4-neighbor-peergroup-holder>
				   </neighbor>
               	   {{else if  ne .neighborAddress  ""}}
				   <neighbor>
					  <af-ipv4-neighbor-address-holder>
						 <af-ipv4-neighbor-address>
							<af-ipv4-neighbor-address>{{.neighborAddress}}</af-ipv4-neighbor-address>
							<activate operation="remove"/>
						 </af-ipv4-neighbor-address>
					  </af-ipv4-neighbor-address-holder>
				   </neighbor>
				   {{end}}
				</default-vrf>
			</ipv4-unicast>
		  </ipv4>
		</address-family>
			   
        </router-bgp>
    </router>
    </routing-system>
</config>  
`

var routerBgpNeighborDelete = `
<config>
       <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <router>
            <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
               <router-bgp-attributes>
                      <neighbor>
                        <neighbor-ips>
                           <neighbor-addr>
                              <router-bgp-neighbor-address>{{.neighbor_address}}</router-bgp-neighbor-address>
                              <associate-peer-group operation="remove"/>
                              <remote-as operation="remove"/>
                           </neighbor-addr>
                        </neighbor-ips>
                     </neighbor>
                  </router-bgp-attributes>
            </router-bgp>
         </router>
      </routing-system>
</config>
`

var configureIPRoute = `
<config>
<ip xmlns="urn:brocade.com:mgmt:brocade-common-def">
	<rtm-config xmlns="urn:brocade.com:mgmt:brocade-rtm">
	<route>
		<static-route-nh>
			<static-route-dest>{{.loopbackIP}}</static-route-dest>
			<static-route-next-hop>{{.VEIP}}</static-route-next-hop>
	</static-route-nh>
	</route>
	</rtm-config>
</ip>
</config>  
`
var deconfigureIPRoute = `
<config>
<ip xmlns="urn:brocade.com:mgmt:brocade-common-def">
	<rtm-config xmlns="urn:brocade.com:mgmt:brocade-rtm">
	<route>
		<static-route-nh operation="remove">
			<static-route-dest>{{.loopbackIP}}</static-route-dest>
			<static-route-next-hop>{{.VEIP}}</static-route-next-hop>
	     </static-route-nh>
	</route>
	</rtm-config>
</ip>
</config>  
`
