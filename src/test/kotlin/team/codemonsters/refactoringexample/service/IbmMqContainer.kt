package team.codemonsters.refactoringexample.service

import org.testcontainers.containers.BindMode
import org.testcontainers.containers.GenericContainer

class IbmMqContainer {

    companion object {
        fun container(): GenericContainer<*> {
            val envVariables = mapOf("LICENSE" to "accept", "MQ_QMGR_NAME" to "QM1")
            return GenericContainer("icr.io/ibm-messaging/mq")
                .withExposedPorts(1414, 1414)
                .withExtraHost("locahost", "0.0.0.0")
                .withEnv(envVariables)
                .withClasspathResourceMapping(
                    "docker/20-config.mqsc",
                    "/etc/mqm/20-config.mqsc",
                    BindMode.READ_ONLY
                )
        }
    }

}