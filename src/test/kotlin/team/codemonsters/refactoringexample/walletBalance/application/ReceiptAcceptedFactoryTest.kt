package team.codemonsters.refactoringexample.walletBalance.application

import com.jayway.jsonpath.JsonPath
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.http.HttpMethod
import org.springframework.messaging.Message
import org.springframework.messaging.MessageHeaders
import org.springframework.messaging.support.MessageBuilder
import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.configuration.RestConfiguration
import team.codemonsters.refactoringexample.util.JsonUtil
import team.codemonsters.refactoringexample.walletBalance.HttpRequestEnvelopeDto
import team.codemonsters.refactoringexample.walletBalance.PsGetBalanceRequest
import team.codemonsters.refactoringexample.walletBalance.UnvalidatedRequest
import team.codemonsters.refactoringexample.walletBalance.ValidRequest
import java.time.Duration
import java.time.OffsetDateTime
import java.util.*

@SpringBootTest
class ReceiptAcceptedFactoryTest(
    @Autowired val restConfiguration: RestConfiguration,
) {

    @Test
    fun `accepted receipt success`() {
        //arrange
        val validRequest = validRequest()
        val config = MqServiceCfg(
            REQ = "req",
            RES = "resp",
            RECEIPT = "receipt",
            timeout = 60000L,
            correlationID = "CorrelationID"
        )

        //act
        val receiptMessageReceived = ReceiptAcceptedFactory.emerge(validRequest, config)

        //assert
        assertReceiptMessage(receiptMessageReceived)
    }

    private fun assertReceiptMessage(receiptAccepted: Result<ReceiptAcceptedFactory>) {
        assertThat(receiptAccepted.isSuccess).isTrue

        val messageId = receiptAccepted.getOrNull()!!.messageId()
        assertThat(messageId).isEqualTo("9f8f6d7e-1645-4c26-837a-ba126906011b")

        val basicJmsRequest = receiptAccepted.getOrNull()!!.basicJmsRequest()

        assertThat(basicJmsRequest.queue).isEqualTo("receipt")

        val payloadContext = JsonPath.parse(basicJmsRequest.payload)
        println(basicJmsRequest.payload)
        assertThat(payloadContext.read<String>("$.requestType")).isEqualTo("operationHistory")
        assertThat(payloadContext.read<String>("$.flag")).isEqualTo("ACCEPTED")
        assertThat(payloadContext.read<String>("$.errorCode")).isNull()
        println(basicJmsRequest.headers)
        assertThat(basicJmsRequest.headers["CorrelationID"]).isEqualTo("9f8f6d7e-1645-4c26-837a-ba126906010d")
        assertThat(basicJmsRequest.headers["X_Method"]).isEqualTo("client-wallet-balance")
    }

    private fun validRequest(): ValidRequest {
        val httpRequestEnvelope =
            createEnvelope(
                correlationId = "correlation-01",
                createValidRequest(),
            )
        val messageHeaders =
            MessageHeaders(
                mapOf(
                    "CorrelationID" to UUID.fromString("9f8f6d7e-1645-4c26-837a-ba126906010d"),
                    "jms_messageId" to UUID.fromString("9f8f6d7e-1645-4c26-837a-ba126906011b"),
                )
            )
        val message: Message<Any> =
            MessageBuilder.createMessage(JsonUtil.toJson(httpRequestEnvelope), messageHeaders)
        val unvalidatedRequest = UnvalidatedRequest(message, restConfiguration, 60000L)

        return ValidRequest.emerge(unvalidatedRequest).getOrNull()!!
    }

    private fun createEnvelope(
        correlationId: String,
        request: PsGetBalanceRequest,
    ) = HttpRequestEnvelopeDto(
        correlationId = correlationId,
        method = HttpMethod.POST.name(),
        service = "api-gateway",
        operation = "client-wallet-balance",
        headers = mapOf("Api-clientId" to "Api-clientId-01", "RequestID" to "request-01"),
        body = JsonUtil.toJson(request),
    )

    private fun createValidRequest() =
        PsGetBalanceRequest(
            requestType = "operationHistory",
            dateFrom = OffsetDateTime.now().minus(Duration.ofSeconds(10)),
            dateTo = OffsetDateTime.now().minus(Duration.ofSeconds(5)),
            operTime =
            OffsetDateTime.now().plus(
                Duration.ofMillis(60000L),
            ).plus(Duration.ofSeconds(20)),
        )

}