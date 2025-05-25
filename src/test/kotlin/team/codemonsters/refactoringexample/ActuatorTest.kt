package team.codemonsters.refactoringexample

import org.assertj.core.api.SoftAssertions
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Tag
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.autoconfigure.web.reactive.AutoConfigureWebTestClient
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.http.HttpMethod
import org.springframework.jms.core.JmsTemplate
import org.springframework.test.context.ActiveProfiles
import org.springframework.test.context.DynamicPropertyRegistry
import org.springframework.test.context.DynamicPropertySource
import org.springframework.test.web.reactive.server.WebTestClient
import team.codemonsters.refactoringexample.configuration.MqConfiguration
import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.mq.*
import team.codemonsters.refactoringexample.mq.enum.ReceiptStatus
import team.codemonsters.refactoringexample.service.IbmMqContainer

@ActiveProfiles("test")
@AutoConfigureWebTestClient
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
class ActuatorTest {

    @Autowired
    private lateinit var webTestClient: WebTestClient

    companion object {
        private val ibmMqContainer = IbmMqContainer.container()

        @JvmStatic
        @DynamicPropertySource
        fun setIbmMqProperties(registry: DynamicPropertyRegistry) {
            registry.add("ibm.mq.connName") { connName() }
        }

        private fun connName() = "${ibmMqContainer.host}(${ibmMqContainer.firstMappedPort})"

        @JvmStatic
        @BeforeAll
        internal fun startContainer() {
            ibmMqContainer.start()
        }

        @JvmStatic
        @AfterAll
        internal fun stopContainer() {
            ibmMqContainer.stop()
        }
    }

    @Test
    fun `health endpoint should return UP status`() {
        webTestClient.get()
            .uri("/actuator/health")
            .exchange()
            .expectStatus().isOk
            .expectBody()
            .consumeWith { response ->
                println("Health response: ${String(response.responseBody!!)}")
            }
            .jsonPath("$.status").isEqualTo("UP")
    }

    @Test
    fun `info endpoint should return application information`() {
        webTestClient.get()
            .uri("/actuator/info")
            .exchange()
            .expectStatus().isOk
            .expectBody()
            .consumeWith { response ->
                println("info response: ${String(response.responseBody!!)}")
            }
            .jsonPath("$.build").exists()
            .jsonPath("$.build.artifact").isEqualTo("mq-rest-sync-adapter")
            .jsonPath("$.build.group").isEqualTo("team.codemonsters")
    }

    @Test
    fun `metrics endpoint should return available metrics`() {
        webTestClient.get()
            .uri("/actuator/metrics")
            .exchange()
            .expectStatus().isOk
            .expectBody()
            .consumeWith { response ->
                println("metrics response: ${String(response.responseBody!!)}")
            }
            .jsonPath("$.names").isArray
    }
} 