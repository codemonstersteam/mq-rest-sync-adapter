package team.codemonsters.refactoringexample.walletBalance

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.http.HttpMethod
import org.springframework.messaging.Message
import org.springframework.messaging.MessageHeaders
import org.springframework.messaging.support.MessageBuilder
import team.codemonsters.refactoringexample.configuration.RestConfiguration
import team.codemonsters.refactoringexample.util.JsonUtil
import java.time.Duration
import java.time.OffsetDateTime
import java.util.UUID

@SpringBootTest
class ValidRequestTest(
    @Autowired val restConfiguration: RestConfiguration,
) {
    @Test
    fun `emerge should return valid request`() {
        // Arrange
        val httpRequestEnvelope =
            createEnvelope(
                correlationId = "correlation-01",
                createValidRequest(),
            )
        val messageHeaders =
            MessageHeaders(mapOf("CorrelationID" to UUID.fromString("9f8f6d7e-1645-4c26-837a-ba126906010d")))
        val message: Message<Any> =
            MessageBuilder.createMessage(JsonUtil.toJson(httpRequestEnvelope), messageHeaders)
        val unvalidatedRequest = UnvalidatedRequest(message, restConfiguration, 60000L)

        // Act
        val result = ValidRequest.emerge(unvalidatedRequest)

        // Assert
        assertThat(result.isSuccess).isTrue
        // Add more assertions based on the expected behavior of the ValidRequest.emerge function
    }

    private fun createEnvelope(
        correlationId: String,
        request: PsGetBalanceRequest,
    ) = HttpRequestEnvelope(
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
