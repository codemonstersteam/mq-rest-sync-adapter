package team.codemonsters.refactoringexample.walletBalance.application

import com.jayway.jsonpath.JsonPath
import org.assertj.core.api.SoftAssertions
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpMethod
import org.springframework.http.HttpStatus
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
import team.codemonsters.refactoringexample.walletBalance.WalletBalance
import team.codemonsters.refactoringexample.walletBalance.infrastructure.RestServiceRequest
import team.codemonsters.refactoringexample.walletBalance.infrastructure.RestServiceResponse
import java.time.Duration
import java.time.OffsetDateTime
import java.util.*

private const val CORRELATION_ID = "9f8f6d7e-1645-4c26-837a-ba126906010d"
private const val MESSAGE_ID = "9f8f6d7e-1645-4c26-837a-ba126906011b"
@SpringBootTest
class TransactionReceiptFactoryTest (
    @Autowired val restConfiguration: RestConfiguration,
) {
    private val requestLifetime = Duration.ofSeconds(10).toMillis()

    @Test
    fun `transaction receipt success`() {
        //arrange
        val validRequest = validRequest()
        val config = MqServiceCfg(
            REQ = "req",
            RES = "resp",
            RECEIPT = "receipt",
            timeout = requestLifetime,
            correlationID = "CorrelationID"
        )
        val restRequest = RestServiceRequest.of(validRequest)

        val response = RestServiceResponse.success(
            request = restRequest,
            httpStatus = HttpStatus.OK,
            headers = HttpHeaders().apply { set("Content-Type", "application/json;charset=UTF-8") },
            body = "{\"balance\": \"100.0\"}",
        )
        val walletBalance = WalletBalance(
            validRequest = validRequest,
            restServiceResponse = response
        )
        //act
        val transactionReceipt = TransactionReceiptFactory(walletBalance, config)
        //assert
        assertReceiptMessage(transactionReceipt)
    }

    private fun assertReceiptMessage(reciept: ReceiptFactory) {
        SoftAssertions.assertSoftly { softly: SoftAssertions ->
            val messageId = reciept.messageId()
            softly.assertThat(messageId).isEqualTo(MESSAGE_ID)

            val basicJmsRequest = reciept.basicJmsRequest()
            softly.assertThat(basicJmsRequest.queue).isEqualTo("receipt")

            val payloadContext = JsonPath.parse(basicJmsRequest.payload)
            println("Payload Context")
            println(basicJmsRequest.payload)
            softly.assertThat(payloadContext.read<String>("$.requestType")).isEqualTo("operationHistory")
            softly.assertThat(payloadContext.read<String>("$.flag")).isEqualTo("SUCCESS")
            softly.assertThat(payloadContext.read<String>("$.errorCode")).isNull()
            softly.assertThat(basicJmsRequest.headers["CorrelationID"]).isEqualTo(CORRELATION_ID)
            softly.assertThat(basicJmsRequest.headers["X_Method"]).isEqualTo("client-wallet-balance")
        }
    }

    private fun validRequest(): ValidRequest {
        val httpRequestEnvelope =
            createEnvelope(
                correlationId = CORRELATION_ID,
                createValidRequest(),
            )
        val messageHeaders =
            MessageHeaders(
                mapOf(
                    "CorrelationID" to UUID.fromString(CORRELATION_ID),
                    "jms_messageId" to UUID.fromString(MESSAGE_ID),
                )
            )
        val message: Message<Any> =
            MessageBuilder.createMessage(JsonUtil.toJson(httpRequestEnvelope), messageHeaders)
        val unvalidatedRequest = UnvalidatedRequest(message, restConfiguration, requestLifetime)

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
            OffsetDateTime.now().minus(Duration.ofMillis(requestLifetime)),
        )

}