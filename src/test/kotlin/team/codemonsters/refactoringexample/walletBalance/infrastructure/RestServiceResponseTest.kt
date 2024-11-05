package team.codemonsters.refactoringexample.walletBalance.infrastructure

import org.assertj.core.api.SoftAssertions
import org.junit.jupiter.api.Test
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpStatus
import team.codemonsters.refactoringexample.walletBalance.HttpRequestEnvelope

internal class RestServiceResponseTest : RestServiceResponseAbstractTest() {

    @Test
    fun `should create success response`() {
        // arrange
        val validRequest = validRequest()
        val restServiceRequest = RestServiceRequest.emerge(validRequest).getOrNull()!!
        val envelope = restServiceRequest.requestEnvelope
        // act
        val response = RestServiceResponse.success(
            request = restServiceRequest,
            httpStatus = HttpStatus.OK,
            headers = HttpHeaders().apply { set("Content-Type", "application/json;charset=UTF-8") },
            body = "{\"balance\": \"100.0\"}",
        )
        // assert
        assertSuccess(response, envelope)
    }

    private fun assertSuccess(response: RestServiceResponse, envelope: HttpRequestEnvelope) {
        SoftAssertions.assertSoftly { softly: SoftAssertions ->
            softly.assertThat(response.serviceBaseUrl).isEqualTo("/api/client/Api-clientId-01/wallet/balance")
            softly.assertThat(response.requestEnvelope).isEqualTo(envelope)
            softly.assertThat(response.httpCode).isEqualTo(HttpStatus.OK)
            softly.assertThat(response.body).isEqualTo("{\"balance\": \"100.0\"}")
            softly.assertThat(response.headers.getFirst("Content-Type")).isEqualTo("application/json;charset=UTF-8")
        }
    }

    @Test
    fun `should create error response`() {
        // arrange
        val validRequest = validRequest()
        val restServiceRequest = RestServiceRequest.emerge(validRequest).getOrNull()!!
        val envelope = restServiceRequest.requestEnvelope
        // act
        val response = RestServiceResponse.error(
            request = restServiceRequest,
            RuntimeException("http connection error"),
        )
        // assert
        assertError(response, envelope)
    }

    private fun assertError(response: RestServiceResponse, envelope: HttpRequestEnvelope) {
        SoftAssertions.assertSoftly { softly: SoftAssertions ->
            softly.assertThat(response.serviceBaseUrl).isEqualTo("/api/client/Api-clientId-01/wallet/balance")
            softly.assertThat(response.requestEnvelope).isEqualTo(envelope)
            softly.assertThat(response.httpCode).isEqualTo(HttpStatus.INTERNAL_SERVER_ERROR)
            softly.assertThat(response.headers).isEmpty()
            softly.assertThat(response.body)
                .isEqualTo("Ошибка http connection error при выполнении запроса к сервису http://base-service-url:8080/api/client/Api-clientId-01/wallet/balance")
        }
    }

    @Test
    fun `should create unsupportedMethod response`() {
        // arrange
        val validRequest = invalidRequest()
        val restServiceRequest = RestServiceRequest.emerge(validRequest).getOrNull()!!
        val envelope = restServiceRequest.requestEnvelope
        // act
        val response = RestServiceResponse.unsupportedMethod(
            request = restServiceRequest,
        )
        // assert
        assertUnsupportedMethod(response, envelope)
    }

    private fun assertUnsupportedMethod(response: RestServiceResponse, envelope: HttpRequestEnvelope) {
        SoftAssertions.assertSoftly { softly: SoftAssertions ->
            softly.assertThat(response.serviceBaseUrl).isEqualTo("/api/client/Api-clientId-01/wallet/balance")
            softly.assertThat(response.requestEnvelope).isEqualTo(envelope)
            softly.assertThat(response.httpCode).isEqualTo(HttpStatus.INTERNAL_SERVER_ERROR)
            softly.assertThat(response.headers).isEmpty()
            softly.assertThat(response.body)
                .isEqualTo("Ошибка при выполнении запроса к сервису http://base-service-url:8080/api/client/Api-clientId-01/wallet/balance: метод DELETE не поддерживается")
        }
    }

}