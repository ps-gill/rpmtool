pipeline {
  agent none
  options {
    timestamps()
  }
  stages {
      stage('base') {
        parallel {
          stage('fedora-40') {
            agent { label 'fedora-40' }
            steps {
              sh '''
              set -euo pipefail
              sudo dnf install golang-bin rpm rpm-devel --assumeyes
              CGO_CFLAGS="-DRPMTOOL_DISABLE_RPMBUILD_MKBUILDDIR -DRPMTOOL_DISABLE_RPMSPEC_NOFINALIZE -DRPMTOOL_DISABLE_BUILDROOTDIR" go build
              '''
            }
          }
          stage('fedora-41') {
            agent { label 'fedora-41' }
            steps {
              sh '''
              set -euo pipefail
              sudo dnf install golang-bin rpm rpm-devel --assumeyes
              go build
              '''
            }
          }
        }
      }
  }
  post {
    failure {
      emailext(
        to: '$DEFAULT_RECIPIENTS',
        subject: '$DEFAULT_SUBJECT',
        body: '$DEFAULT_CONTENT',
      )
    }
    fixed {
      emailext(
        to: '$DEFAULT_RECIPIENTS',
        subject: '$DEFAULT_SUBJECT',
        body: '$DEFAULT_CONTENT',
      )
    }
  }
}
