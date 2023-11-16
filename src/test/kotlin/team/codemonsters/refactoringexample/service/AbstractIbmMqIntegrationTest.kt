package team.codemonsters.refactoringexample.service

import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Tag
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.test.context.DynamicPropertyRegistry
import org.springframework.test.context.DynamicPropertySource
import org.testcontainers.containers.BindMode
import org.testcontainers.containers.GenericContainer

@Tag("integration-test")
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
abstract class AbstractIbmMqIntegrationTest {

    companion object {
        val envVariables = mapOf("LICENSE" to "accept", "MQ_QMGR_NAME" to "QM1")
        val container = GenericContainer("icr.io/ibm-messaging/mq")
            .withExposedPorts(1414, 1414)
            .withExtraHost("locahost", "0.0.0.0")
            .withEnv(envVariables)
            .withClasspathResourceMapping(
                "docker/20-config.mqsc",
                "/etc/mqm/20-config.mqsc",
                BindMode.READ_ONLY
            )

        @JvmStatic
        @DynamicPropertySource
        fun setIbmMqProperties(registry: DynamicPropertyRegistry) {
            registry.add("ibm.mq.connName", Companion::connName)
        }

        private fun connName() = "${container.host}(${container.firstMappedPort})"

        @JvmStatic
        @BeforeAll
        internal fun setUp() {
            container.start()
        }

    }

}