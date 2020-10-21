pipeline {
    agent { docker 'ubuntu_golang_1_11' }
    environment {
          HOME_DIR = "$WORKSPACE"
          REPORT_DIR = "$HOME_DIR/reports"
          GOPATH = "$HOME_DIR"
          TEST_DIR= "$HOME_DIR/src/efa-server/test"
    }
    stages {
        stage('cleanup') {
            steps {
                 sh '''
                   rm -rf  $HOME_DIR/bin/*
                   rm -rf  $REPORT_DIR
                 '''
            }
        }
        stage('build-server') {
            steps {
                sh '''
                    #Setting PATH variable in environment section not working
                    #So setting here
                    # sudo apt-get install -y upx
                    export PATH=$PATH:/usr/local/go/bin
                    ls -ltr $HOME_DIR/src/efa-server
                    cd $HOME_DIR/src/efa-server
                    go vet $(go list ./... | grep -v generated)
                    BUILD_STAMP=`date +%y-%m-%d:%H:%M:%S`
                    VERSION=2.0.0
                    export LDFLAG="-s -w -X efa/infra.Version=$VERSION -X efa/infra.BuildStamp=$BUILD_STAMP"
                    echo $LDFLAG
                    go install -ldflags "$LDFLAG"
                    upx $HOME_DIR/bin/efa-server
                    ls -lh  $HOME_DIR/bin/efa-server
                '''
            }
        }
        stage('build-client') {
            steps {
                sh '''
                    #Setting PATH variable in environment section not working
                    #So setting here
                    export PATH=$PATH:/usr/local/go/bin
                    cd $HOME_DIR/src/efa
                    go vet $(go list ./... | grep -v generated)
                    BUILD_STAMP=`date +%y-%m-%d:%H:%M:%S`
                    VERSION=2.0.0
                    export LDFLAG="-s -w -X efa/infra.Version=$VERSION -X efa/infra.BuildStamp=$BUILD_STAMP"
                    echo $LDFLAG
                    go install -ldflags "$LDFLAG"
                    upx $HOME_DIR/bin/efa
                    ls -lh  $HOME_DIR/bin/efa
                '''
            }
        }
        stage('Test Prepare'){
              steps{
                  sh '''
                     #Get Test Report Generator
                     export PATH=$PATH:/usr/local/go/bin
                     go clean -testcache
                  '''
              }
        }
        stage('Run Test') {
            steps {
                sh '''
                    ln -s $HOME_DIR /EFA
                    # bash /EFA/scripts/docker_test.sh

                '''
            }

        }


    }
    post {
        always {
            archiveArtifacts artifacts: 'bin/efa*, fingerprint: true
        }
     }
}

