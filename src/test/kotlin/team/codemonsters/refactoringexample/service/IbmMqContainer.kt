package team.codemonsters.refactoringexample.service

import org.testcontainers.containers.BindMode
import org.testcontainers.containers.GenericContainer

class IbmMqContainer {

    companion object {
        const val IMAGE = "ibmcom/mq:9.2.4.0-r1-amd64"
        fun container(): GenericContainer<*> {
            val envVariables = mapOf("LICENSE" to "accept", "MQ_QMGR_NAME" to "QM1")
            return GenericContainer(IMAGE)
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