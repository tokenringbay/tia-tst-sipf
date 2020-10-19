package base

var overlayGatewayCreate = `
<config>
<overlay-gateway xmlns="urn:brocade.com:mgmt:brocade-tunnels">
     <name>{{.gw_name}}</name>
     <gw-type>{{.gw_type}}</gw-type>
     <ip>
        <interface>
           <loopback>
              <loopback-id>{{.loopback_id}}</loopback-id>
           </loopback>
        </interface>
     </ip>
     {{if  eq .map_vni_auto  "true" }}
     <map>
        <vlan-and-bd>
           <vni>
              <auto></auto>
           </vni>
        </vlan-and-bd>
     </map>
     {{end}}
     <activate></activate>
</overlay-gateway>
</config>
`

var overlayGatewayDelete = `
<config>
<overlay-gateway xmlns="urn:brocade.com:mgmt:brocade-tunnels" operation="remove">
     <name>{{.gw_name}}</name>
</overlay-gateway>
</config>
`

var evpnInstanceCreate = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <evpn-config xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <evpn>
               <evpn-instance>
                  <instance-name>{{.evi_name}}</instance-name>
                  <route-target>
                     <both>
                        <target-community>auto</target-community>
                        <ignore-as></ignore-as>
                     </both>
                  </route-target>
                  <route-distinguisher>
                     <auto></auto>
                  </route-distinguisher>
                  <duplicate-mac-timer>
                     <duplicate-mac-timer-value>{{.duplicate_mac_timer}}</duplicate-mac-timer-value>
                     <max-count>{{.duplicate_mac_timer_max_count}}</max-count>
                  </duplicate-mac-timer>
               </evpn-instance>
            </evpn>
         </evpn-config>
      </routing-system>
</config>
`

var evpnInstanceDelete = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <evpn-config xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <evpn operation="delete">
               <evpn-instance>
                  <instance-name>{{.evi_name}}</instance-name>
               </evpn-instance>
            </evpn>
         </evpn-config>
      </routing-system>
</config>
`

var interfaceNumberedCreate = `
<config>
    <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <{{.int_type}}>
            <name>{{.int_name}}</name>
			<description>{{.description}}</description>
            <ip>
               <ip-config xmlns="urn:brocade.com:mgmt:brocade-ip-config">
                  <address>
                     <address>{{.ipaddress}}</address>
                  </address>
               </ip-config>
            </ip>
         </{{.int_type}}>
      </interface>
</config>      
`

var interfaceUnnumberedCreate = `
<config>
    <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <{{.int_type}}>
            <name>{{.int_name}}</name>
            <ip>
               <ip-config xmlns="urn:brocade.com:mgmt:brocade-ip-config">
                  <unnumbered>
                      <ip-donor-interface-type>{{.donor_interface_type}}</ip-donor-interface-type>
                      <ip-donor-interface-name>{{.donor_interface_name}}</ip-donor-interface-name>
                  </unnumbered>
               </ip-config>
            </ip>
         </{{.int_type}}>
      </interface>
</config>      
`

var interfaceUnnumberedDelete = `
<config>
    <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <{{.int_type}}>
            <name>{{.int_name}}</name>
            <ip>
               <ip-config xmlns="urn:brocade.com:mgmt:brocade-ip-config">
                  <unnumbered operation="remove"/>
               </ip-config>
            </ip> 
			<bfd>
               <interval operation="remove"/>
            </bfd>
			<description operation="remove"> </description>
         </{{.int_type}}>
      </interface>
</config>      
`

var interfaceNumberedDelete = `
<config>
    <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <{{.int_type}}>
            <name>{{.int_name}}</name>
            <ip>
               <ip-config xmlns="urn:brocade.com:mgmt:brocade-ip-config">
                  <address operation="remove">
                     <address>{{.ipaddress}}</address>
                  </address>
               </ip-config>
            </ip>
            <bfd>
               <interval operation="remove">
               </interval>
            </bfd>
            <description operation="remove"> </description>
         </{{.int_type}}>
      </interface>
</config>   
`

var interfaceDescDelete = `
<config>
    <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <{{.int_type}}>
            <name>{{.int_name}}</name>
            <description operation = "remove"> </description>
         </{{.int_type}}>
      </interface>
</config>   
`
var interfaceSpeedDelete = `
<config>
    <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <{{.int_type}}>
            <name>{{.int_name}}</name>
            <speed operation="remove"></speed> 
         </{{.int_type}}>
      </interface>
</config>   
`
var interfaceActivate = `
<config>
    <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <{{.int_type}}>
            <name>{{.int_name}}</name>
             <shutdown operation="remove"></shutdown>
         </{{.int_type}}>
      </interface>
</config>      
`

var bgpRouterCreate = `
<config>
       <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <router>
            <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
               <router-bgp-attributes>
                  <local-as>{{.local_as}}</local-as>
                  <capability>
                     <as4-enable></as4-enable>
                  </capability>
                  {{if eq .bfd_enable "Yes"}}
				  <bfd>
					   <interval>
						  <min-tx>{{.bfd_min_tx}}</min-tx>
						  <min-rx>{{.bfd_min_rx}}</min-rx>
						  <multiplier>{{.bfd_multiplier}}</multiplier>
					   </interval> 	
                  </bfd>
                  {{end}}
                  <fast-external-fallover></fast-external-fallover>
                  <neighbor>
                     <peer-grps>
                        <neighbor-peer-grp>
                           <router-bgp-neighbor-peer-grp>{{.peer_group_name}}</router-bgp-neighbor-peer-grp>
                           <peer-group-name></peer-group-name>
                           <description>{{.description}}</description>
                           {{if eq .bfd_enable "Yes"}}
                           <bfd>
                              <bfd-enable></bfd-enable>
                           </bfd>
                           {{end}}
                        </neighbor-peer-grp>
                     </peer-grps>
                  </neighbor>
               </router-bgp-attributes>
               <address-family>
                  <ipv4>
                     <ipv4-unicast>
                        <default-vrf>
                           <default-vrf-selected></default-vrf-selected>

							{{if eq .detrisibuteConnected "Yes"}}
						    <af-ipv4-uc-and-vrf-cmds-call-point-holder>
                              <redistribute>
                                 <connected>
                                    <redistribute-connected></redistribute-connected>
                                 </connected>
                              </redistribute>
                           </af-ipv4-uc-and-vrf-cmds-call-point-holder>
                           {{else if eq .detrisibuteConnectedWithRouteMap "Yes"}}
						    <af-ipv4-uc-and-vrf-cmds-call-point-holder>
                              <redistribute>
                                 <connected>
                                    <redistribute-connected></redistribute-connected>
                                    <redistribute-route-map>ToR-map</redistribute-route-map>
                                 </connected>
                              </redistribute>
                           </af-ipv4-uc-and-vrf-cmds-call-point-holder>
                           {{else}} 
                             {{if  ne .network_address  ""}}
                              <network>
                                <network-ipv4-address>{{.network_address}}</network-ipv4-address>
                              </network>
                             {{end}}
                           {{end}} 
                           <af-common-cmds-holder>
                              <maximum-paths>
                                 <load-sharing-value>{{.max_paths}}</load-sharing-value>
                              </maximum-paths>
                              <graceful-restart>
			                     <graceful-restart-status>
                                 </graceful-restart-status>
			                   </graceful-restart>
                           </af-common-cmds-holder>
                        </default-vrf>
                     </ipv4-unicast>
                  </ipv4>    
                  {{if  eq .evpn  "Yes" }}
                  <l2vpn>
                     <evpn>
                        {{if eq .retain_route_target_all "Yes" }}
                        <retain>
                           <route-target>
                              <all></all>
                           </route-target>
                        </retain> 
                        {{end}} 
                        <neighbor>
                           <evpn-peer-group>
                              <evpn-neighbor-peergroup-name>{{.peer_group_name}}</evpn-neighbor-peergroup-name>
                              <encapsulation>vxlan</encapsulation>
			      {{if eq .next_hop_unchanged "Yes" }}
			      <next-hop-unchanged></next-hop-unchanged>
                              {{end}} 
                               {{if  ne .allowas_in  "0"}}
                              <allowas-in>{{.allowas_in}}</allowas-in>
                              {{end}}
								<enable-peer-as-check></enable-peer-as-check>
                              <activate></activate>
                           </evpn-peer-group>
                        </neighbor>
                        <graceful-restart>
			               <graceful-restart-status>
                           </graceful-restart-status>
			            </graceful-restart>
                     </evpn>
                  </l2vpn>
                  {{end}}            
               </address-family>
            </router-bgp>
         </router>
      </routing-system>
</config>  
`

var routerIDCreate = `
<config>
	  <ip xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <rtm-config xmlns="urn:brocade.com:mgmt:brocade-rtm">
            <router-id>{{.router_id}}</router-id>
         </rtm-config>
      </ip>
</config>      
`

var routerIDDelete = `
<config>
	<ip xmlns="urn:brocade.com:mgmt:brocade-common-def">
		<rtm-config xmlns="urn:brocade.com:mgmt:brocade-rtm">
			<router-id operation="remove"/>
		</rtm-config>
	</ip>
</config>   
`

var routerBgpNeighborCreate = `
<config>
       <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <router>
            <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
               <router-bgp-attributes>
                      <neighbor>   
                        {{if  eq .is_leaf  "Yes" }}  
                        <peer-grps>
                          <neighbor-peer-grp>
                           <router-bgp-neighbor-peer-grp>{{.peer_group_name}}</router-bgp-neighbor-peer-grp>
                           <peer-group-name></peer-group-name>
                           <remote-as>{{.remote_as}}</remote-as>
                          </neighbor-peer-grp>
                        </peer-grps>
                         {{else}}
                        <neighbor-ips>
                           <neighbor-addr>
                              <router-bgp-neighbor-address>{{.neighbor_address}}</router-bgp-neighbor-address>
                              <remote-as>{{.remote_as}}</remote-as>
                           </neighbor-addr>
                        </neighbor-ips>
                        {{end}}
                     </neighbor>
                  </router-bgp-attributes>
               
            </router-bgp>
         </router>
      </routing-system>
</config>  
`

var routerBgpNeighborAssociateWithPeerGroup = `
<config>
       <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <router>
            <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
               <router-bgp-attributes>
                      <neighbor>
                        <neighbor-ips>
                           <neighbor-addr>
                              <router-bgp-neighbor-address>{{.neighbor_address}}</router-bgp-neighbor-address>
                              <associate-peer-group>{{.peer_group_name}}</associate-peer-group>
                           </neighbor-addr>
                        </neighbor-ips>
                     </neighbor>
                  </router-bgp-attributes>
            </router-bgp>
         </router>
      </routing-system>
</config>
`
var routerBgpNeighborMultihop = `
<config>
       <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <router>
            <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
               <router-bgp-attributes>
                      <neighbor>
                        <neighbor-ips>
                           <neighbor-addr>
                              <router-bgp-neighbor-address>{{.neighbor_address}}</router-bgp-neighbor-address>
                             <ebgp-multihop>
                               <ebgp-multihop-count>{{.bgp_multihop}}</ebgp-multihop-count>
                           	 </ebgp-multihop>
		       	    {{if eq .next_hop_self true}}
				<next-hop-self>
					<next-hop-self-status/>
				</next-hop-self>
		 	   {{end}}
                           </neighbor-addr>
                        </neighbor-ips>
                     </neighbor>
                  </router-bgp-attributes>
            </router-bgp>
         </router>
      </routing-system>
</config>
`

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
					{{if eq .bfdEnabled "Yes"}}
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
				   {{end}}
			           {{if  ne .neighborAddress  ""}}
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

var routerBgpDelete = `
<config>	      
     <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
        <router>
           <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp" operation="remove">
           </router-bgp>
        </router>
     </routing-system>
 </config>  
`

var ipLoopbackCreate = `
<config>  
	  <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
          <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
            <loopback xmlns="urn:brocade.com:mgmt:brocade-intf-loopback">
               <id>{{.loopback_id}}</id>
               <ip xmlns="urn:brocade.com:mgmt:brocade-ip-config">
                  <ip-config>
                     <address>
                        <address>{{.ipaddress}}</address>
                     </address>
                  </ip-config>
               </ip>
            </loopback>
         </interface>
      </routing-system>
</config>
`

var ipLoopbackDelete = `
<config>
    <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
        <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
            <loopback operation="remove" xmlns="urn:brocade.com:mgmt:brocade-intf-loopback">
                <id>{{.loopback_id}}</id>
            </loopback>
        </interface>
    </routing-system>
</config>
`

var ipLoopbackActivate = `
<config>  
	  <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
          <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
            <loopback xmlns="urn:brocade.com:mgmt:brocade-intf-loopback">
               <id>{{.loopback_id}}</id>
               <shutdown operation="remove"></shutdown>
            </loopback>
         </interface>
      </routing-system>
</config>
`

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
         <ipv6>
            <static-ag-ipv6-config xmlns="urn:brocade.com:mgmt:brocade-vrrp">
               <anycast-gateway-mac>
                  <ipv6-anycast-gateway-mac>{{.ipv6_anycast_gateway_mac}}</ipv6-anycast-gateway-mac>
               </anycast-gateway-mac>
            </static-ag-ipv6-config>
         </ipv6>
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
         <ipv6>
            <static-ag-ipv6-config xmlns="urn:brocade.com:mgmt:brocade-vrrp">
               <anycast-gateway-mac>
                  <ipv6-anycast-gateway-mac operation="remove"></ipv6-anycast-gateway-mac>
               </anycast-gateway-mac>
            </static-ag-ipv6-config>
         </ipv6>
      </routing-system>
</config>
`

var nodeIDClusterPriorityCreate = `
<config>
   <node-id xmlns="urn:brocade.com:mgmt:brocade-node">
      <node-id>{{.node_id}}</node-id>
      {{if  eq .configure_cluster_priority  "true" }}
      <cluster xmlns="http://brocade.com/ns/brocade-cluster">
          <management>
             <principal-priority>{{.cluster_priority}}</principal-priority>
          </management>
      </cluster>
      {{end}}
   </node-id>  
</config>
`
var nodeIDClusterPriorityDelete = `
<config>
   <node-id xmlns="urn:brocade.com:mgmt:brocade-node">
      <node-id>{{.node_id}}</node-id>
      <cluster xmlns="http://brocade.com/ns/brocade-cluster" operation="delete">    
      </cluster>
   </node-id>  
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

var intVeDeleteIP = `
  <config>
     <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
            <ve>
               <name>{{.name}}</name>
               <ip xmlns="urn:brocade.com:mgmt:brocade-ip-config">
                  <ip-config>
                     <address operation="remove">
                        <address>{{.ip_address}}</address>
                     </address>
                  </ip-config>
               </ip>
            </ve>
         </interface>
    </routing-system>
  </config>
`

var intVeSetIP = `
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
            </ve>
         </interface>
    </routing-system>
  </config>
`

var intVeDelete = `
  <config>
     <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
            <ve operation="remove">
               <name>{{.name}}</name>
            </ve>
         </interface>
    </routing-system>
  </config>
`

var mctClusterCreate = `
<config>
  <cluster xmlns="urn:brocade.com:mgmt:brocade-mct">
         <cluster-name>{{.cluster_name}}</cluster-name>
         <cluster-id>{{.cluster_id}}</cluster-id>
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

var mctClusterUndeploy = `
<config>
	<cluster xmlns="urn:brocade.com:mgmt:brocade-mct">
		<cluster-name>{{.cluster_name}}</cluster-name>
		<cluster-id>{{.cluster_id}}</cluster-id>
		<deploy operation="remove"></deploy>
	</cluster>
</config>
`

var mctClusterRemovePeerIP = `
<config>
	<cluster xmlns="urn:brocade.com:mgmt:brocade-mct">
		<cluster-name>{{.cluster_name}}</cluster-name>
		<cluster-id>{{.cluster_id}}</cluster-id>
		<peer operation="delete">
			<peer-ip>{{.peer_ip}}</peer-ip>
		</peer>
	</cluster>
</config>
`

var mctClusterAddPeerIP = `
<config>
	<cluster xmlns="urn:brocade.com:mgmt:brocade-mct">
		<cluster-name>{{.cluster_name}}</cluster-name>
		<cluster-id>{{.cluster_id}}</cluster-id>
		<peer>
			<peer-ip>{{.peer_ip}}</peer-ip>
		</peer>
		<deploy></deploy>
	</cluster>
</config>
`

var mctControlVlanCreate = `
<config>
	<interface-vlan xmlns="urn:brocade.com:mgmt:brocade-interface">
         <vlan>
            <name>{{.control_vlan}}</name>
            <router-interface>
               <ve-config>{{.control_ve}}</ve-config>
            </router-interface>
            <description>{{.description}}</description>
         </vlan>
      </interface-vlan>
</config>
`
var mctControlVeDelete = `
<config>
	<interface-vlan xmlns="urn:brocade.com:mgmt:brocade-interface">
         <vlan>
            <name>{{.control_vlan}}</name>
            <router-interface operation="remove">
               <ve-config>{{.control_ve}}</ve-config>
            </router-interface>
         </vlan>
      </interface-vlan>
</config>
`
var mctControlVlanDelete = `
<config>
	<interface-vlan xmlns="urn:brocade.com:mgmt:brocade-interface">
         <vlan operation="remove">
            <name>{{.control_vlan}}</name>
         </vlan>
      </interface-vlan>
</config>
`

var intPortChannelCreate = `
<config>
	<interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <port-channel>
            <name>{{.name}}</name>
         </port-channel>
      </interface>
</config>
`

var intPortChannelDeactivate = `
<config>
        <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <port-channel>
            <name>{{.name}}</name>
            <shutdown></shutdown>
         </port-channel>
      </interface>
</config>
`

var intPortChannelActivate = `
<config>
	<interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <port-channel>
            <name>{{.name}}</name>
            <shutdown operation="remove"> </shutdown>
         </port-channel>
      </interface>
</config>
`

var intPortChannelSpeedSet = `
<config>
	<interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <port-channel>
            <name>{{.name}}</name>
             <po-speed>{{.speed}}</po-speed>
         </port-channel>
      </interface>
</config>
`
var intPortChannelDescriptionSet = `
<config>
	<interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <port-channel>
            <name>{{.name}}</name>
             <description>{{.description}}</description>
         </port-channel>
      </interface>
</config>
`
var intPortChannelDelete = `
<config>
	<interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <port-channel operation="remove">
            <name>{{.name}}</name>
         </port-channel>
      </interface>
</config>
`

var intAddToPortChannel = `
<config>
	 <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <ethernet>
            <name>{{.name}}</name>
            <description>{{.port_channel_description}}</description>
            <channel-group>
               <port-int>{{.port_channel}}</port-int>
               <mode>{{.port_channel_mode}}</mode>
               <type>{{.port_channel_type}}</type>
            </channel-group>
         </ethernet>
      </interface>
</config>
`
var intPhySpeed = `
<config>
         <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <ethernet>
            <name>{{.name}}</name>
            <speed>{{.speed}}</speed>
         </ethernet>
      </interface>
</config>
`

var intRemoveFromPortChannel = `
<config>
	 <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <ethernet>
            <name>{{.name}}</name>
            <description operation="remove"></description>
            <channel-group operation="remove">
               <port-int>{{.port_channel}}</port-int>
            </channel-group>
         </ethernet>
      </interface>
</config>
`

var intPhyActivate = `
<config>
	 <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <ethernet>
            <name>{{.name}}</name>
			<shutdown operation="remove"> </shutdown>
         </ethernet>
      </interface>
</config>
`
var intPhyDeactivate = `
<config>
         <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
         <ethernet>
            <name>{{.intf_name}}</name>
	    <shutdown></shutdown>
         </ethernet>
      </interface>
</config>
`

var intEnableInterfaces = `
<config>
	 <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
		 {{range .interface_names}}
         <ethernet>
            <name>{{.}}</name>
            <shutdown operation="remove"> </shutdown>
         </ethernet>
         {{end}}
      </interface>
</config>
`
var intDisableInterfaces = `
<config>
	 <interface xmlns="urn:brocade.com:mgmt:brocade-interface">
		 {{range .interface_names}}
         <ethernet>
            <name>{{.}}</name>
            <shutdown> </shutdown>
         </ethernet>
         {{end}}
      </interface>
</config>
`

var persistConfig = `
<bna-config-cmd xmlns="urn:brocade.com:mgmt:brocade-ras">
      <src>running-config</src>
      <dest>startup-config</dest>
   </bna-config-cmd>
`

var configureSystemWideL2Mtu = `
<config>
	<global-mtu-conf xmlns="urn:brocade.com:mgmt:brocade-interface">
		<global-l2-mtu>{{.l2_mtu_value}}</global-l2-mtu>
	</global-mtu-conf>
</config>
`

var unconfigureSystemWideL2Mtu = `
<config>
	<global-mtu-conf xmlns="urn:brocade.com:mgmt:brocade-interface">
		<global-l2-mtu operation="remove"/>
	</global-mtu-conf>
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

var configureSystemWideIPv6Mtu = `
<config>
	<global-mtu-conf xmlns="urn:brocade.com:mgmt:brocade-interface">
		<ipv6 xmlns="urn:brocade.com:mgmt:brocade-ipv6-config">
			<global-ipv6-mtu>{{.ip_mtu_value}}</global-ipv6-mtu>
		</ipv6>
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

var unconfigureSystemWideIPv6Mtu = `
<config>
	<global-mtu-conf xmlns="urn:brocade.com:mgmt:brocade-interface">
		<ipv6 xmlns="urn:brocade.com:mgmt:brocade-ipv6-config">
			<global-ipv6-mtu operation="remove"/>
		</ipv6>
	</global-mtu-conf>
</config>
`

var routeMapPermitCreate = `
<config>
	  <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <route-map xmlns="urn:brocade.com:mgmt:brocade-ip-policy">
            <name>ToR-map</name>
            <action-rm>permit</action-rm>
            <instance>20</instance>
         </route-map>
      </routing-system>
</config>      
`
var routeMapDenyCreate = `
<config>
	  <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <route-map xmlns="urn:brocade.com:mgmt:brocade-ip-policy">
            <name>ToR-map</name>
            <action-rm>deny</action-rm>
            <instance>10</instance>
         </route-map>
      </routing-system>
</config>      
`
var routeMapDenyIPPrefixCreate = `
<config>
	  <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <route-map xmlns="urn:brocade.com:mgmt:brocade-ip-policy">
            <name>ToR-map</name>
            <action-rm>deny</action-rm>
            <instance>10</instance>
            <content>
               <match>
                  <ip>
                     <address>
                        <prefix-list-rmm>fabric_links_ip</prefix-list-rmm>
                     </address>
                  </ip>
               </match>
            </content>
         </route-map>
      </routing-system>
</config>      
`
var routeMapIPPrefixCreate = `
<config>
	  <ip xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <hide-prefix-holder xmlns="urn:brocade.com:mgmt:brocade-ip-policy">
            <prefix-list>
               <name>fabric_links_ip</name>
               <seq-keyword>seq</seq-keyword>
               <instance>10</instance>
               <action-ipp>permit</action-ipp>
               <prefix-ipp>0.0.0.0/0</prefix-ipp>
               <ge-ipp>31</ge-ipp>
               <le-ipp>31</le-ipp>
            </prefix-list>
         </hide-prefix-holder>
      </ip>
</config>      
`

var routeMapIPPrefixDelete = `
<config>
	  <ip xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <hide-prefix-holder xmlns="urn:brocade.com:mgmt:brocade-ip-policy">
            <prefix-list operation="remove">
               <name>fabric_links_ip</name>
               <seq-keyword>seq</seq-keyword>
               <instance>10</instance>
            </prefix-list>
         </hide-prefix-holder>
      </ip>
</config>      
`
var routeMapPermitDelete = `
<config>
	  <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <route-map xmlns="urn:brocade.com:mgmt:brocade-ip-policy" operation="remove">
            <name>ToR-map</name>
            <action-rm>permit</action-rm>
            <instance>20</instance>
         </route-map>
      </routing-system>
</config>      
`
var routeMapDenyDelete = `
<config>
	  <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
         <route-map xmlns="urn:brocade.com:mgmt:brocade-ip-policy" operation="remove">
            <name>ToR-map</name>
            <action-rm>deny</action-rm>
            <instance>10</instance>
         </route-map>
      </routing-system>
</config>      
`

var bgpRouterNonClosCreate = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <router-bgp-attributes>
               <local-as>{{.local_as}}</local-as>
               <capability>
                  <as4-enable></as4-enable>
               </capability>
           {{if eq .eBGPBFD true}}
               <bfd>
		  <interval>          
		  <min-tx>{{.bfd_min_tx}}</min-tx>
		  <min-rx>{{.bfd_min_rx}}</min-rx>
		  <multiplier>{{.bfd_multiplier}}</multiplier>
		  </interval> 	
               </bfd>
          {{end}}
               <fast-external-fallover></fast-external-fallover>
            </router-bgp-attributes>
            <address-family>
               <ipv4>
                  <ipv4-unicast>
                     <default-vrf>
                        <af-common-cmds-holder>
                           <maximum-paths>
                              <load-sharing-value>{{.max_paths}}</load-sharing-value>
                           </maximum-paths>
                        </af-common-cmds-holder>
                     </default-vrf>
                  </ipv4-unicast>
               </ipv4>
            </address-family>
         </router-bgp>
      </router>
   </routing-system>
</config>  
`

var bgpRouterNonClosNetworkCreate = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <address-family>
               <ipv4>
                  <ipv4-unicast>
                        <default-vrf>
                           <default-vrf-selected></default-vrf-selected>
                             {{if  ne .network_address  ""}}
                              <network>
                                <network-ipv4-address>{{.network_address}}</network-ipv4-address>
                              </network>
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

var bgpRouterNonClosEVPNPeerGroupCreate = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <router-bgp-attributes>
               <neighbor>
                  <peer-grps>
                     <neighbor-peer-grp>
                        <router-bgp-neighbor-peer-grp>{{.EVPNPeerGroup}}</router-bgp-neighbor-peer-grp>
                        <peer-group-name></peer-group-name>
                        <description>{{.EVPNPeerGroupDescription}}</description>
                        <ebgp-multihop>
                           <ebgp-multihop-count>{{.EVPNPeerGroupMultiHop}}</ebgp-multihop-count>
                        </ebgp-multihop>
                     </neighbor-peer-grp>
                  </peer-grps>
               </neighbor>
            </router-bgp-attributes>
         </router-bgp>
      </router>
   </routing-system>
</config>
`

var bgpRouterNonClosEBGPPeerGroupCreate = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <router-bgp-attributes>
               <neighbor>
               <peer-grps>
                  <neighbor-peer-grp>
                     <router-bgp-neighbor-peer-grp>{{.eBGPPeerGroup}}</router-bgp-neighbor-peer-grp>
                     <peer-group-name></peer-group-name>
                     <description>{{.eBGPPeerGroupDescription}}</description>
                      {{if eq .eBGPBFD true}}
					      <bfd>
						      <bfd-enable></bfd-enable>
                     </bfd>
                     {{end}}
                  </neighbor-peer-grp>
               </peer-grps>
               </neighbor>
            </router-bgp-attributes>
         </router-bgp>
      </router>
   </routing-system>
</config>
`
var bgpRouterNonClosEVPNNeighborCreate = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <router-bgp-attributes>
               <neighbor>
                  <neighbor-ips>
                     <neighbor-addr>
                        <router-bgp-neighbor-address>{{.neighborAddress}}</router-bgp-neighbor-address>
                        <remote-as>{{.remote_as}}</remote-as>                                               
                     </neighbor-addr>
                  </neighbor-ips>
               </neighbor>
            </router-bgp-attributes>            
         </router-bgp>
      </router>
   </routing-system>
</config>
`

var bgpRouterNonClosEVPNNeighborAssociate = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <router-bgp-attributes>
               <neighbor>
                  <neighbor-ips>
                     <neighbor-addr>
                        <router-bgp-neighbor-address>{{.neighborAddress}}</router-bgp-neighbor-address>                        
                        <associate-peer-group>{{.peer_group_name}}</associate-peer-group>
                        <update-source>
			   <loopback>{{.loopback_number}}</loopback>
                        </update-source>                         
                     </neighbor-addr>
                  </neighbor-ips>
               </neighbor>
            </router-bgp-attributes>            
         </router-bgp>
      </router>
   </routing-system>
</config>
`

/*var bgpRouterNonClosEVPNNeighborUpdateSource = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <router-bgp-attributes>
               <neighbor>
                  <neighbor-ips>
                     <neighbor-addr>
                        <router-bgp-neighbor-address>{{.neighborAddress}}</router-bgp-neighbor-address>

                     </neighbor-addr>
                  </neighbor-ips>
               </neighbor>
            </router-bgp-attributes>
         </router-bgp>
      </router>
   </routing-system>
</config>
`*/

var bgpRouterNonClosEVPNNeighborProperties = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <address-family>
               <l2vpn>
                  <evpn>
                  {{if eq .retain_rt_all true}}
                  <retain>
                     <route-target>
                        <all></all>
                     </route-target>
                  </retain> 
                  {{end}} 
                  <neighbor>
                     <evpn-peer-group>
                        <evpn-neighbor-peergroup-name>{{.EVPNPeerGroup}}</evpn-neighbor-peergroup-name>
                        <encapsulation>{{.encap}}</encapsulation>
                        {{if eq .next_hop_unchanged true }}
                           <next-hop-unchanged></next-hop-unchanged>
                        {{end}}
                        <activate></activate>
                     </evpn-peer-group>
                  </neighbor>
                  <graceful-restart>
                     <graceful-restart-status>
                     </graceful-restart-status>
                  </graceful-restart>
                  </evpn>
               </l2vpn>
            </address-family>
         </router-bgp>
      </router>
   </routing-system>
</config>
`

var bgpRouterNonClosBGPNeighborCreate = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <router-bgp-attributes>
               <neighbor>
               <neighbor-ips>
                  <neighbor-addr>
                     <router-bgp-neighbor-address>{{.neighbor_address}}</router-bgp-neighbor-address>
                     <remote-as>{{.remote_as}}</remote-as>
                  </neighbor-addr>
               </neighbor-ips>
               </neighbor>
            </router-bgp-attributes>
         </router-bgp>
      </router>
   </routing-system>
</config>
`
var bgpRouterNonClosBGPNeighborAssociate = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <router-bgp-attributes>
               <neighbor>
               <neighbor-ips>
                  <neighbor-addr>
                     <router-bgp-neighbor-address>{{.neighbor_address}}</router-bgp-neighbor-address>
                     <remote-as>{{.remote_as}}</remote-as>
                     {{if ne .peer_group_name ""}}
                     <associate-peer-group>{{.peer_group_name}}</associate-peer-group>
                     {{end}}
                  </neighbor-addr>
               </neighbor-ips>
               </neighbor>
            </router-bgp-attributes>
         </router-bgp>
      </router>
   </routing-system>
</config>
`
var bgpRouterNonClosBGPNeighborProperties = `
<config>
   <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
      <router>
         <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
            <router-bgp-attributes>
               <neighbor>
               <neighbor-ips>
                  <neighbor-addr>
                     <router-bgp-neighbor-address>{{.neighbor_address}}</router-bgp-neighbor-address>
                     <remote-as>{{.remote_as}}</remote-as>
                     {{if eq .next_hop_self true}}
                     <next-hop-self>
                        <next-hop-self-status/>
                     </next-hop-self>
                     {{end}}
                     {{if eq .bfd_enabled true}}
		     <bfd>
		        <bfd-enable></bfd-enable>
                     </bfd>
                     {{end}}
                  </neighbor-addr>
               </neighbor-ips>
               </neighbor>
            </router-bgp-attributes>
         </router-bgp>
      </router>
   </routing-system>
</config>
`
