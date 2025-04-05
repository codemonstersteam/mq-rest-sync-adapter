package team.codemonsters.refactoringexample.walletBalance.infrastructure

import org.assertj.core.api.SoftAssertions
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
import java.util.*

internal class RestServiceRequestTest {

    private val restConfiguration: RestConfiguration = RestConfiguration(
        configs = mapOf(
            "api-gateway" to RestServiceCfg(
                url = "http://base-service-url:8080",
                basicAuth = null,
                operations = mapOf("client-wallet-balance" to "/api/client/\${Api-clientId}/wallet/balance")
            )
        )
    )

    private val CORRELATION_ID = "9f8f6d7e-1645-4c26-837a-ba126906010d"
    private val MESSAGE_ID = "9f8f6d7e-1645-4c26-837a-ba126906011b"
    private val requestLifetime = Duration.ofSeconds(10).toMillis()

    @Test
    fun `should create RestServiceRequest with valid request`() {
        // arrange
        val validRequest = validRequest()
        // act
        val restServiceRequest = RestServiceRequest.of(validRequest)
        // assert
        assertRequest(restServiceRequest, validRequest)
    }

    private fun assertRequest(
        request: RestServiceRequest,
        validRequest: ValidRequest
    ) {
        SoftAssertions.assertSoftly { softly: SoftAssertions ->
            //1 какой http метод запроса будет использован
            softly.assertThat(request.method).isEqualTo(HttpMethod.POST.name())
            // 2 на какой url будет отправлен запрос
            softly.assertThat(request.uri.toString())
                .isEqualTo("http://base-service-url:8080/api/client/Api-clientId-01/wallet/balance")
            // 3 что будет содержаться в теле запроса
            softly.assertThat(request.body).isEqualTo(validRequest.httpRequestEnvelope.body)
            // 4 какие заголовки будут использованы в запросе
            softly.assertThat(request.headers().getFirst("Content-Type")).isEqualTo("application/json;charset=UTF-8")
            softly.assertThat(request.headers().getFirst("Api-clientId")).isEqualTo("Api-clientId-01")
            softly.assertThat(request.headers().getFirst("RequestID")).isEqualTo("request-01")
        }
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