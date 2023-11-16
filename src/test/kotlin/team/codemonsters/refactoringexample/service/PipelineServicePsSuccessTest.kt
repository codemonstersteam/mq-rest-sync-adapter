package team.codemonsters.refactoringexample.service

import com.maciejwalkowiak.wiremock.spring.ConfigureWireMock
import com.maciejwalkowiak.wiremock.spring.EnableWireMock
import org.assertj.core.api.SoftAssertions
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Tag
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.http.HttpMethod
import org.springframework.jms.core.JmsTemplate
import org.springframework.test.context.ActiveProfiles
import org.springframework.test.context.DynamicPropertyRegistry
import org.springframework.test.context.DynamicPropertySource
import team.codemonsters.refactoringexample.configuration.MqConfiguration
import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.mq.*
import team.codemonsters.refactoringexample.mq.enum.ReceiptStatus
import team.codemonsters.refactoringexample.util.JsonUtil
import java.time.Duration
import java.time.OffsetDateTime


@ActiveProfiles("test")
@Tag("integration-test")
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@EnableWireMock(ConfigureWireMock(name = "rest-api-gateway", property = "rest.configs.api-gateway.url"))
class PipelineServicePsSuccessTest(
    @Autowired val jmsTemplate: JmsTemplate,
    @Autowired val jmsPublisher: MqPublisher,
    @Autowired private val mqConfiguration: MqConfiguration,
) {
    private val mqConfig: MqServiceCfg = mqConfiguration.configs["ps"]!!

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
    @Tag("PipelineServicePS")
    fun `successful wallet balance request`() {
        // Arrange
        val requestQueue = "IN.QUEUE.PS"
        val message = createRequestEnvelope(requestQueue)
        val originalId = "01"
        val responseQueue = "OUT.QUEUE.PS"
        val receiptQueue = "RECEIPT.QUEUE.PS"
        // Act
        publishToQueue(message, originalId)
        // Assert
        assertResponsesInQueues(receiptQueue, responseQueue)
    }

    private fun assertResponsesInQueues(receiptQueue: String, responseQueue: String) {
        assertReceiptSuccess(receiptQueue)
        assertReceiptAccepted(receiptQueue)
        assertResponse(responseQueue)
    }

    private fun assertReceiptSuccess(receiptQueue: String) {
        val receiptString = jmsTemplate.receiveAndConvert(receiptQueue).toString()
        val receipt = JsonUtil.fromString(receiptString, ReceiptPs::class.java)
        SoftAssertions.assertSoftly { softly: SoftAssertions ->
            softly.assertThat(receipt.flag).isEqualTo(ReceiptStatus.SUCCESS)
            softly.assertThat(receipt.description).isNull()
        }
    }

    private fun assertReceiptAccepted(receiptQueue: String) {
        val receiptString = jmsTemplate.receiveAndConvert(receiptQueue).toString()
        val receipt = JsonUtil.fromString(receiptString, ReceiptPs::class.java)
        SoftAssertions.assertSoftly { softly: SoftAssertions ->
            softly.assertThat(receipt.flag).isEqualTo(ReceiptStatus.ACCEPTED)
            softly.assertThat(receipt.errorCode).isNull()
        }
    }

    private fun assertResponse(responseQueue: String) {
        val responseString = jmsTemplate.receiveAndConvert(responseQueue).toString()
        val responseEnvelope = JsonUtil.fromString(responseString, HttpResponseEnvelope::class.java)
        assertThatResponse(responseEnvelope)
    }

    private fun assertThatResponse(response: HttpResponseEnvelope) {
        SoftAssertions.assertSoftly { softly: SoftAssertions ->
            softly.assertThat(response.correlationId).isEqualTo("correlationId")
            softly.assertThat(response.status).isEqualTo(200)
            softly.assertThat(response.message).isEqualTo("OK")
            softly.assertThat(response.method).isEqualTo("POST")
            softly.assertThat(response.service).isEqualTo("api-gateway")
            softly.assertThat(response.operation).isEqualTo("client-wallet-balance")
            softly.assertThat(response.body)
                .contains("\"data\":{\"createTime\":1600443296227,\"operationId\":\"01\",\"balance\":{\"value\":999.9,\"currency\":\"rub\"}}")
        }
    }

    private fun createRequestEnvelope(queueName: String): BasicJmsRequest<String> {
        val jmsHeaders: Map<String, String> = createJmsHeaders()
        val request = createValidRequest()
        val correlationId = "correlationId"
        val envelope = createEnvelope(correlationId, request)
        val payloadJsonString = JsonUtil.toJson(envelope)
        return BasicJmsRequest(
            correlationId = correlationId,
            queue = queueName,
            payload = payloadJsonString,
            headers = jmsHeaders
        )
    }

    private fun createEnvelope(
        correlationId: String,
        request: PsGetBalanceRequest
    ) = HttpRequestEnvelope.emerge(
        correlationId = correlationId,
        httpMethod = HttpMethod.POST.name(),
        service = "api-gateway",
        operation = "client-wallet-balance",
        httpHeaders = mapOf("Api-clientId" to "Api-clientId-01", "RequestID" to "request-01"),
        httpBody = JsonUtil.toJson(request),
    ).getOrThrow()

    private fun createValidRequest() = PsGetBalanceRequest(
        requestType = "operationHistory",
        dateFrom = OffsetDateTime.now().minus(Duration.ofSeconds(10)),
        dateTo = OffsetDateTime.now().minus(Duration.ofSeconds(5)),
        operTime = OffsetDateTime.now().plus(
            Duration.ofMillis(mqConfig.timeout)
        ).plus(Duration.ofSeconds(20))
    )

    private fun createJmsHeaders() = mapOf(
        "X_Method" to "client-wallet-balance",
        "jms_messageId" to "de741c66-4eea-4316-a42c-a42bb15c8e32",
        "X_TransactionID" to "transaction-01",
        "X_RequestID" to "x-request-01",
        "RequestID" to "request-01",
    )

    private fun publishToQueue(message: BasicJmsRequest<String>, originalId: String) {
        jmsPublisher.sendToMq(
            message, originalMsgId = originalId
        )
    }
}