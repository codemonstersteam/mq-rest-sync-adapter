package team.codemonsters.refactoringexample.walletBalance

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.springframework.http.HttpMethod
import team.codemonsters.refactoringexample.configuration.RestBasicAuth
import team.codemonsters.refactoringexample.configuration.RestConfiguration
import team.codemonsters.refactoringexample.configuration.RestServiceCfg
import team.codemonsters.refactoringexample.mq.PsGetBalanceRequest
import team.codemonsters.refactoringexample.util.JsonUtil
import java.time.Duration
import java.time.OffsetDateTime

class HttpRequestEnvelopeTestFail {

    @Test
    fun `should create invalid envelope when body is null`() {
        //arrange
        val envelopeDto = createInvalidEnvelopeDto("request-01", null)
        val restConfiguration = restConfigurationInvalidService()
        //act
        val resultEnvelope = HttpRequestEnvelope.emerge(
            envelopeDto,
            restConfiguration
        )
        //assert
        assertThat(resultEnvelope.isFailure)
        assertThat(resultEnvelope.exceptionOrNull()!!.message).isEqualTo("Отсутствует тела запроса в конверте")
    }

    @Test
    fun `should create invalid envelope when body is blank`() {
        //arrange
        val envelopeDto = createInvalidEnvelopeDto("request-01", " ")
        val restConfiguration = restConfigurationInvalidService()
        //act
        val resultEnvelope = HttpRequestEnvelope.emerge(
            envelopeDto,
            restConfiguration
        )
        //assert
        assertThat(resultEnvelope.isFailure)
        assertThat(resultEnvelope.exceptionOrNull()!!.message).isEqualTo("Отсутствует тела запроса в конверте")
    }


    @Test
    fun `should create invalid envelope when service not found`() {
        //arrange
        val envelopeDto = createEnvelopeDto("request-01", createValidBalanceRequest())
        val restConfiguration = restConfigurationInvalidService()
        //act
        val resultEnvelope = HttpRequestEnvelope.emerge(
            envelopeDto,
            restConfiguration
        )
        //assert
        assertThat(resultEnvelope.isFailure)
        assertThat(resultEnvelope.exceptionOrNull()!!.message).isEqualTo("Не найдена конфигурация для сервиса: api-gateway")
    }

    private fun createInvalidEnvelopeDto(correlationId: String, body: String?)
    = HttpRequestEnvelopeDto(
        correlationId = correlationId,
        method = HttpMethod.POST.name(),
        service = "api-gateway",
        operation = "client-wallet-balance",
        headers = mapOf("Api-clientId" to "Api-clientId-01", "RequestID" to "request-01"),
        body = body,
    )

    private fun createEnvelopeDto(
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

    private fun createValidBalanceRequest() =
        PsGetBalanceRequest(
            requestType = "operationHistory",
            dateFrom = OffsetDateTime.now().minus(Duration.ofSeconds(10)),
            dateTo = OffsetDateTime.now().minus(Duration.ofSeconds(5)),
            operTime =
            OffsetDateTime.now().plus(
                Duration.ofMillis(60000L),
            ).plus(Duration.ofSeconds(20)),
        )

    private fun restConfiguration() = RestConfiguration(
        configs = mapOf(
            "api-gateway" to RestServiceCfg(
                url = "http://localhost:8080",
                basicAuth = RestBasicAuth("test", "test"),
                operations = mapOf("client-wallet-balance" to "/api/client/\${Api-clientId}/wallet/balance")
            )
        )
    )

    private fun restConfigurationInvalidService() = RestConfiguration(
        configs = mapOf(
            "api-gateway-invalid" to RestServiceCfg(
                url = "http://localhost:8080",
                basicAuth = RestBasicAuth("test", "test"),
                operations = mapOf("client-wallet-balance" to "/api/client/\${Api-clientId}/wallet/balance")
            )
        )
    )

}