package team.codemonsters.refactoringexample.walletBalance.infrastructure
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.springframework.http.HttpMethod
import org.springframework.messaging.Message
import org.springframework.messaging.MessageHeaders
import org.springframework.messaging.support.MessageBuilder
import team.codemonsters.refactoringexample.configuration.RestConfiguration
import team.codemonsters.refactoringexample.configuration.RestServiceCfg
import team.codemonsters.refactoringexample.util.JsonUtil
import team.codemonsters.refactoringexample.walletBalance.HttpRequestEnvelopeDto
import team.codemonsters.refactoringexample.walletBalance.PsGetBalanceRequest
import team.codemonsters.refactoringexample.walletBalance.UnvalidatedRequest
import team.codemonsters.refactoringexample.walletBalance.ValidRequest
import java.time.Duration
import java.time.OffsetDateTime
import java.util.UUID

internal class RestServiceRequestFailTest {

    private val restConfiguration: RestConfiguration = RestConfiguration(
        configs = mapOf(
            "api-gateway" to RestServiceCfg(
                url = "\\8080",
                basicAuth = null,
                operations = mapOf("client-wallet-balance" to "/api/client/\${Api-clientId}/wallet/balance")
            )
        )
    )

    private val CORRELATION_ID = "9f8f6d7e-1645-4c26-837a-ba126906010d"
    private val MESSAGE_ID = "9f8f6d7e-1645-4c26-837a-ba126906011b"


    private val requestLifetime = Duration.ofSeconds(10).toMillis()

    @Test
    fun `should fail emerge RestServiceRequest with uri in config`() {
        // given
        val validRequest = validRequest()

        // when
        val restServiceRequest = RestServiceRequest.emerge(validRequest)

        // then
        assertThat(restServiceRequest.isFailure).isTrue
        assertThat(restServiceRequest.exceptionOrNull()!!.message)
            .contains("Failed to create RestServiceRequest cause the url {\\8080/api/client/Api-clientId-01/wallet/balance} is not valid")
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