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

class HttpRequestEnvelopeTestSuccess {

    @Test
    fun `should create valid envelope`() {
        //arrange
        val envelopeDto = envelopeDto("request-01", validBalanceRequest())
        val restConfiguration = restConfiguration()
        //act
        val resultEnvelope = HttpRequestEnvelope.emerge(
            envelopeDto,
            restConfiguration
        )
        //assert
        assertThat(resultEnvelope.isSuccess)
    }

    private fun envelopeDto(
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

    private fun validBalanceRequest() =
        PsGetBalanceRequest(
            requestType = "operationHistory",
            dateFrom = OffsetDateTime.now().minus(Duration.ofSeconds(10)),
            dateTo = OffsetDateTime.now().minus(Duration.ofSeconds(5)),
            operTime =
            OffsetDateTime.now().plus(Duration.ofSeconds(26)),
        )

    private fun restConfiguration()
    = RestConfiguration(
        configs = mapOf(
            "api-gateway" to RestServiceCfg(
                url = "http://localhost:8080",
                basicAuth = RestBasicAuth("test", "test"),
                operations = mapOf("client-wallet-balance" to "/api/client/\${Api-clientId}/wallet/balance")
            )
        )
    )

}