.. code:: robotframework

    *** Test Cases ***


    IPFABRIC FABRIC SHOW
		[Setup]          Run Keywords  START_REST_SERVER
        Log To Console   ${APP}
        ${result}=       Run Process  ${APP}  fabric  show
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}  | Link IP Range             | 10.10.10.0/23   |
        Should Contain   ${op} 	| Loopback IP Range         | 172.32.254.0/24 |
        Should Contain   ${op} 	| Loopback Port Number      | 1               |
        Should Contain   ${op} 	| VTEP Loopback Port Number | 2               |
        Should Contain   ${op} 	| Spine ASN Block           | 64512           |
        Should Contain   ${op} 	| LEAF ASN Block            | 65000-65534     |
        [Teardown]       Run Keywords  DELETE_DB
	
    IPFABRIC ADD SPINE AND LEAVES
		[Setup]          Run Keywords  START_REST_SERVER
        ${result}=       Run Process  ${APP}  configure  add  --spines  ${SWITCH_1}  --leaves  ${SWITCH_2},${SWITCH_3}
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}  Configure Fabric [Success]
        [Teardown]       Run Keywords  DELETE_DB
     
    
    ##################      Validation Tests #########################################
    
    IPFABRIC ADD ONLY SPINE
		[Setup]          Run Keywords  START_REST_SERVER
        ${result}=       Run Process  ${APP}  configure  add  --spines  ${SWITCH_1} 
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}  ${VALIDATE_FABRIC_FAILED}
        Should Contain   ${op}  ${VALIDATE_NO_LEAF}
        [Teardown]       Run Keywords  DELETE_DB
    
    IPFABRIC ADD ONLY LEAVES
		[Setup]          Run Keywords  START_REST_SERVER
        ${result}=       Run Process  ${APP}  configure  add  --leaves  ${SWITCH_2},${SWITCH_3} 
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}  ${VALIDATE_FABRIC_FAILED}
        Should Contain   ${op}  ${VALIDATE_NO_SPINE}
        [Teardown]       Run Keywords  DELETE_DB    
    
    IPFABRIC ADD SPINE THEN LEAVES
		[Setup]          Run Keywords  START_REST_SERVER
        ${result}=       Run Process  ${APP}  configure  add  --spines  ${SWITCH_1} 
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}  ${VALIDATE_FABRIC_FAILED}
        Should Contain   ${op}  ${VALIDATE_NO_LEAF}
        ${result}=       Run Process  ${APP}  configure  add  --leaves  ${SWITCH_2},${SWITCH_3} 
        ${op}=           Get Variable Value  ${result.stdout}
        Should Contain   ${op}  Configure Fabric [Success]
        Log To Console   ${op}
        [Teardown]       Run Keywords  DELETE_DB    
     
    IPFABRIC ADD LEAVES THEN SPINE
		[Setup]          Run Keywords  START_REST_SERVER
        ${result}=       Run Process  ${APP}  configure  add  --leaves  ${SWITCH_2},${SWITCH_3}   
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}  ${VALIDATE_FABRIC_FAILED}
        Should Contain   ${op}  ${VALIDATE_NO_SPINE}
        ${result}=       Run Process  ${APP}  configure  add  --spines  ${SWITCH_1}
        ${op}=           Get Variable Value  ${result.stdout}
        Should Contain   ${op}  Configure Fabric [Success]
        Log To Console   ${op}
        [Teardown]       Run Keywords  DELETE_DB
    
    *** Keywords ***
    START_REST_SERVER
    	${result}=       Start Process  ${APP}  restart

    EVALUATE_APP
    	${system}=    Evaluate    platform.system()    platform
        log to console    \nI am running on ${system}
        Set Global Variable  ${APP}  ${LINUX_APP}
        Run Keyword If 	 '${system}' == "MSYS_NT-6.1"  Set Global Variable  ${APP}  ${WINDOWS_APP}
        
    *** Keywords ***
    DELETE_DB
    	${result}=       Run Process  ${APP}  stop
    	${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        ${result}=       BuiltIn.Sleep  2
        ${result}=       Run Process  rm  /c/var/efa/efa.db
        ${result}=       Run Process  rm  /var/efa/efa.db
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
       
    *** Settings ***
    Suite Setup         Run Keywords  EVALUATE_APP
    Library             OperatingSystem
    Library             Process
    Library             BuiltIn
    Variables           001_DCFabric_SLX.yaml
 
