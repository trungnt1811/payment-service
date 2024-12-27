#!/usr/bin/env groovy

import groovy.json.JsonOutput
import java.text.SimpleDateFormat
import java.util.Optional
import hudson.tasks.test.AbstractTestResultAction
import hudson.model.Actionable
import hudson.tasks.junit.CaseResult
import groovy.json.StringEscapeUtils
import hudson.Util

def dockerRegistryProtocol = "https"
def dockerRegistryHostname = "741499881865.dkr.ecr.ap-southeast-1.amazonaws.com"
def dockerRegistry = "${dockerRegistryProtocol}://${dockerRegistryHostname}"
def slackNotificationChannel = 'jenkins'
def slackReleaseChannel = 'releases'
def awsCredentialsId = 'aws.genomicdao.iam.jenkins'
def notifySlack(text, channel, attachments) {
    def slackURL = 'https://hooks.slack.com/services/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'
    def jenkinsIcon = 'https://wiki.jenkins-ci.org/download/attachments/2916393/logo.png'

    def payload = JsonOutput.toJson([text: text,
        channel: channel,
        username: "Jenkins",
        icon_url: jenkinsIcon,
        attachments: attachments
    ])

    sh "curl -X POST -H 'Content-type: application/json' --data '${payload}' ${slackURL} --max-time 5"
}

def slackToken
def uploadToSlack(message, channel, fileName, filePath) {
    sh "curl https://slack.com/api/files.upload -F token='${slackToken}' -F channels='${channel}' -F initial_comment='${message}' -F title='${fileName}' -F filename='${fileName}' -F file=@'${filePath}'"
}

def developServerBranch = "";
def developServerEnvFile = "";
def stagingServerBranch = "";
def stagingServerEnvFile = "";
def productionServerBranch = "";
def productionServerEnvFile = "";

def loadBuildEnv = {
  load ".env.build"
  developServerBranch = "${DEVELOP_BRANCH}"
  developServerEnvFile = "${DEVELOP_ENV_FILE}"
  developServerSSHConfigName = "${DEVELOP_SSH_CONFIG_NAME}"
//   stagingServerBranch = ${STAGING_BRANCH}
//   stagingServerEnvFile = ${STAGING_ENV_FILE}
  productionServerBranch = "${PRODUCTION_BRANCH}"
  productionServerEnvFile = "${PRODUCTION_ENV_FILE}"
  productionServerSSHConfigName = "${PRODUCTION_SSH_CONFIG_NAME}"
}

def isDeploymentBranch = { ->
    return env.BRANCH_NAME == developServerBranch || env.BRANCH_NAME == productionServerBranch
}

def isResultGoodForDeployment = { ->
    return currentBuild.result == null
}

def shouldDeployToDevelop = { ->
    return env.BRANCH_NAME == developServerBranch
}

def shouldDeployToProduction = { ->
    return env.BRANCH_NAME == productionServerBranch
}

def commitHash = "";
def getCommitHash = {
    commitHash = sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
}

def author = "";
def getGitAuthor = {
    author = sh(returnStdout: true, script: "git --no-pager show -s --format='%an' ${commitHash}").trim()
}
def commitMessage = "";
def getLastCommitMessage = {
    commitMessage = StringEscapeUtils.escapeJavaScript(sh(returnStdout: true, script: 'git log -1 --pretty=%B').trim())
}

def version = "";
def getGitVersion = {
    version = sh(returnStdout: true, script: 'git describe --always').trim()
}

def populateGlobalVariables = {
    getCommitHash()
    getLastCommitMessage()
    getGitAuthor()
    getGitVersion()
}

def image, container
def buildName = "${env.BRANCH_NAME.replaceAll('[^a-zA-Z0-9]+','-').toLowerCase()}"
def buildNumber = "${env.BUILD_NUMBER}"
def jobName = "${env.JOB_NAME}"
jobName = "${jobName.getAt(0..(jobName.indexOf('/') - 1))}/${env.BRANCH_NAME}"
onchainHandlerImageName = "${dockerRegistryHostname}/life-ai/life-point-onchain-handler:${buildName}-${env.BUILD_ID}"

def buildEnvFileId = 'life-ai.life-point.build-env'

def dateFormatter = new SimpleDateFormat("yyyy-MM-dd-HHmmss")
def dateString = dateFormatter.format(new Date())
def zipFileName = "life-ai.life-point.${buildName}-${buildNumber}.deployment.${dateString}"
def zipFileNameWithExtension = "${zipFileName}.zip"

def testSummary = ""
def total = 0
def failed = 0
def skipped = 0

@NonCPS
def getTestSummary = { ->
    def testResultAction = currentBuild.rawBuild.getAction(AbstractTestResultAction.class)
    def summary = ""

    if (testResultAction != null) {
        total = testResultAction.getTotalCount()
        failed = testResultAction.getFailCount()
        skipped = testResultAction.getSkipCount()

        summary = "Passed: " + (total - failed - skipped)
        summary = summary + (", Failed: " + failed)
        summary = summary + (", Skipped: " + skipped)
    } else {
        summary = "No tests found"
    }
    return summary
}

@NonCPS
def getFailedTests = { ->
    def testResultAction = currentBuild.rawBuild.getAction(AbstractTestResultAction.class)
    def failedTestsString = "```"

    if (testResultAction != null) {
        def failedTests = testResultAction.getFailedTests()

        if (failedTests.size() > 9) {
            failedTests = failedTests.subList(0, 8)
        }

        for(CaseResult cr : failedTests) {
            failedTestsString = failedTestsString + "${cr.getFullDisplayName()}:\n${cr.getErrorStackTrace()}\n\n"
        }
        failedTestsString = failedTestsString + "```"
    }
    return failedTestsString
}

throttle(['life-ai/life-point']) {
    node {
        properties([disableConcurrentBuilds()])

        try {
            stage('Clone repository') {
                /* Let's make sure we have the repository cloned to our workspace */
                checkout scm
                populateGlobalVariables()
            }

            stage('Prepare') {
                configFileProvider([
                    configFile(fileId: buildEnvFileId, targetLocation: '.env.build'),
                ]){
                  loadBuildEnv()
                }
            }

            stage('Build') {
                image = docker.build(onchainHandlerImageName, "--pull -f ./onchain-handler/Dockerfile ./onchain-handler")
            }

            stage('Test') {
                echo "Tests passed"
            }

            stage('Push images') {
                withAWS(region: 'ap-southeast-1', credentials: awsCredentialsId) {
                  def loginToECR = ecrLogin()
                  sh loginToECR
                  image.push()

                  image.push("${buildName}")

                  image.push("${commitHash}")

                  if (env.BRANCH_NAME == 'master') {
                    image.push("latest")

                    image.push("${version}")
                  }
                }
            }

            if(isDeploymentBranch() && isResultGoodForDeployment()){
                stage('Deploy') {
                    if (shouldDeployToProduction()) {
                        echo "shouldDeployToProduction"
                    } else if (shouldDeployToDevelop()) {
                        echo "shouldDeployToDevelop"

                        // Prepare
                        configFileProvider([
                            configFile(fileId: developServerEnvFile, targetLocation: 'deployment/.env')
                        ]) {
                          dir("${env.WORKSPACE}/deployment/scripts"){
                              sh "bashly generate"
                              sh "chmod -R +x ."
                              sh "./cli backup pack -f ${zipFileName} --include-env"
                          }
                        }

                        // Deploy
                        sshCommand = """ls -al \
                        && cd deployment \
                        && ./scripts/cli service stop \
                        && ./scripts/cli backup clean -d ../backups \
                        && ./scripts/cli backup create -d ../backups \
                        && cd .. \
                        && rm -rf deployment \
                        && unzip -d deployment ${zipFileNameWithExtension} \
                        && cd deployment \
                        && chmod +x ./scripts/cli \
                        && ./scripts/cli deploy"""

                        sshPublisher(
                          failOnError: true,
                          publishers: [
                            sshPublisherDesc(
                              configName: "${developServerSSHConfigName}",
                              verbose: true,
                              transfers: [
                                sshTransfer(
                                  sourceFiles: "deployment/archives/${zipFileNameWithExtension}",
                                  remoteDirectory: ".",
                                  removePrefix: "deployment/archives/",
                                  execCommand: sshCommand,
                                  usePty: true
                                )
                              ]
                            )
                          ]
                        )

                        // Clean up
                        sh "rm -rf packages/deployment/archives"
                    }
                }
            }

            stage('Report') {
                def buildColor = currentBuild.result == null ? "good" : "warning"
                def buildStatus = currentBuild.result == null ? "succeeded" : currentBuild.result
                def duration = "${Util.getTimeSpanString(System.currentTimeMillis() - currentBuild.startTimeInMillis)}"
                testSummary = getTestSummary()

                if (failed > 0) {
                    def failedTestsString = getFailedTests()
                    notifySlack("", slackNotificationChannel, [
                        [
                            title: "${jobName}, build #${env.BUILD_NUMBER}",
                            title_link: "${env.RUN_DISPLAY_URL}",
                            color: "${buildColor}",
                            author_name: "${author}",
                            text: "Job *${jobName}, build #${env.BUILD_NUMBER}* ${buildStatus} after ${duration}",
                            fields: [
                                [
                                    title: "Branch",
                                    value: "${env.BRANCH_NAME}",
                                    short: true
                                ],
                                [
                                    title: "Last Commit",
                                    value: "${commitMessage}",
                                    short: true
                                ],
                                [
                                    title: "Test Results",
                                    value: "${testSummary}",
                                    short: true
                                ]
                            ]
                        ],
                        [
                            title: "Failed Tests",
                            color: "${buildColor}",
                            text: "${failedTestsString}",
                            "mrkdwn_in": ["text"],
                        ]
                  ])
                } else {
                    notifySlack("", slackNotificationChannel, [
                        [
                            title: "${jobName}, build #${env.BUILD_NUMBER}",
                            title_link: "${env.RUN_DISPLAY_URL}",
                            color: "${buildColor}",
                            author_name: "${author}",
                            text: "Job *${jobName}, build #${env.BUILD_NUMBER}* ${buildStatus} after ${duration}",
                            fields: [
                                [
                                    title: "Branch",
                                    value: "${env.BRANCH_NAME}",
                                    short: true
                                ],
                                [
                                    title: "Last Commit",
                                    value: "${commitMessage}",
                                    short: true
                                ],
                                [
                                    title: "Test Results",
                                    value: "${testSummary}",
                                    short: true
                                ]
                            ]
                        ]
                    ])
                }
            }
        } catch (e) {
            def buildStatus = "failed"
            def duration = "${Util.getTimeSpanString(System.currentTimeMillis() - currentBuild.startTimeInMillis)}"
            testSummary = getTestSummary()
            if (failed > 0) {
                def failedTestsString = getFailedTests()
                notifySlack("", slackNotificationChannel, [
                    [
                        title: "${jobName}, build #${env.BUILD_NUMBER}",
                        title_link: "${env.RUN_DISPLAY_URL}",
                        color: "danger",
                        author_name: "${author}",
                        text: "Job *${jobName}, build #${env.BUILD_NUMBER}* ${buildStatus} after ${duration}",
                        fields: [
                            [
                                title: "Branch",
                                value: "${env.BRANCH_NAME}",
                                short: true
                            ],
                            [
                                title: "Last Commit",
                                value: "${commitMessage}",
                                short: true
                            ],
                            [
                                title: "Test Results",
                                value: "${testSummary}",
                                short: true
                            ],
                            [
                                title: "Error",
                                value: "${e}",
                                short: true
                            ]
                        ]
                    ],
                    [
                        title: "Failed Tests",
                        color: "${buildColor}",
                        text: "${failedTestsString}",
                        "mrkdwn_in": ["text"],
                    ]
                ])
            } else {
                notifySlack("", slackNotificationChannel, [
                    [
                        title: "${jobName}, build #${env.BUILD_NUMBER}",
                        title_link: "${env.RUN_DISPLAY_URL}",
                        color: "danger",
                        author_name: "${author}",
                        text: "Job *${jobName}, build #${env.BUILD_NUMBER}* ${buildStatus} after ${duration}",
                        fields: [
                            [
                                title: "Branch",
                                value: "${env.BRANCH_NAME}",
                                short: true
                            ],
                            [
                                title: "Last Commit",
                                value: "${commitMessage}",
                                short: true
                            ],
                            [
                                title: "Test Results",
                                value: "${testSummary}",
                                short: true
                            ],
                            [
                                title: "Error",
                                value: "${e}",
                                short: true
                            ]
                        ]
                    ]
                ])
            }
            throw e
        }
    }
}