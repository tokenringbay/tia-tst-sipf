.. code:: robotframework

    *** Test Cases ***
    #Test cases for CLI validation


    ESA TOP LEVEL Commands
		[Setup]          Run Keywords  START_REST_SERVER
        Log To Console   ${APP}
        ${result}=       Run Process  ${APP}
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}  configure   Configure IP Fabric
        Should Contain   ${op}  execution   Execution
        Should Contain   ${op}  fabric      Fabric Commands
        Should Contain   ${op}  help        Help about any command
        Should Contain   ${op}  rest        rest Server Commands
        [Teardown]       Run Keywords  DELETE_DB
	
    ESA Configure Command
        [Setup]          Run Keywords  START_REST_SERVER
        Log To Console   ${APP}
        ${result}=       Run Process  ${APP}  configure  --help
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}    add         Add switches to the fabric.
        Should Contain   ${op}    delete      Delete switches from the fabric.
        [Teardown]       Run Keywords  DELETE_DB

    ESA Configure Add Command
        [Setup]          Run Keywords  START_REST_SERVER
        Log To Console   ${APP}
        ${result}=       Run Process  ${APP}  configure  add  --help
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}         --force           Force the configuration on the devices
        Should Contain   ${op}     -h, --help            help for add
        Should Contain   ${op}     --leaves string   Comma seperated list of spine IP Address/Hostnames
        Should Contain   ${op}     --spines string   Comma seperated list of spine IP Address/Hostnames
        Should Contain   ${op}     --user string     Username for the list of devices
        Should Contain   ${op}     --pass string     Password for the list of devices
        [Teardown]       Run Keywords  DELETE_DB

    ESA Configure Add Command User Flag
        [Setup]          Run Keywords  START_REST_SERVER
        Log To Console   ${APP}
        ${result}=       Run Process  ${APP}  configure  add  --user  adm
        ${op}=           Get Variable Value  ${result.stderr}
        Log To Console   ${op}
        Should contain   ${op}      Error: Required both flags "user" and "pass".
        Should Contain   ${op}         --force           Force the configuration on the devices
        Should Contain   ${op}     -h, --help            help for add
        Should Contain   ${op}     --leaves string   Comma seperated list of spine IP Address/Hostnames
        Should Contain   ${op}     --spines string   Comma seperated list of spine IP Address/Hostnames
        Should Contain   ${op}     --user string     Username for the list of devices
        Should Contain   ${op}     --pass string     Password for the list of devices
        [Teardown]       Run Keywords  DELETE_DB

    ESA Configure Add Command Pass Flag
        [Setup]          Run Keywords  START_REST_SERVER
        Log To Console   ${APP}
        ${result}=       Run Process  ${APP}  configure  add  --pass  adm
        ${op}=           Get Variable Value  ${result.stderr}
        Log To Console   ${op}
        Should contain   ${op}      Error: Required both flags "user" and "pass".
        Should Contain   ${op}         --force           Force the configuration on the devices
        Should Contain   ${op}     -h, --help            help for add
        Should Contain   ${op}     --leaves string   Comma seperated list of spine IP Address/Hostnames
        Should Contain   ${op}     --spines string   Comma seperated list of spine IP Address/Hostnames
        Should Contain   ${op}     --user string     Username for the list of devices
        Should Contain   ${op}     --pass string     Password for the list of devices
        [Teardown]       Run Keywords  DELETE_DB

    ESA Execution Command
        [Setup]          Run Keywords  START_REST_SERVER
        Log To Console   ${APP}
        ${result}=       Run Process  ${APP}  execution  --help
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}     list        Display the list of executions
        [Teardown]       Run Keywords  DELETE_DB

    ESA Fabric Command
        [Setup]          Run Keywords  START_REST_SERVER
        Log To Console   ${APP}
        ${result}=       Run Process  ${APP}  fabric  --help
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}   show        show fabric properties.
        Should Contain   ${op}   update      Update fabric properties.
        [Teardown]       Run Keywords  DELETE_DB

    ESA Rest Command
        [Setup]          Run Keywords  START_REST_SERVER
        Log To Console   ${APP}
        ${result}=       Run Process  ${APP}  rest  --help
        ${op}=           Get Variable Value  ${result.stdout}
        Log To Console   ${op}
        Should Contain   ${op}   start       Start the Rest Server.
        Should Contain   ${op}   stop        Stop the Rest Server.
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
 
