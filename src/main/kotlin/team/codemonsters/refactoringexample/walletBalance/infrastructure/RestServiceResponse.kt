package team.codemonsters.refactoringexample.walletBalance.infrastructure

import org.springframework.http.HttpHeaders
import org.springframework.http.HttpStatus
import team.codemonsters.refactoringexample.walletBalance.HttpRequestEnvelope

data class RestServiceResponse private constructor(
    val serviceBaseUrl: String,
    val requestEnvelope: HttpRequestEnvelope,
    val httpCode: HttpStatus,
    val headers: HttpHeaders,
    val body: String?
) {

    companion object {
        fun success(
            request: RestServiceRequest,
            httpStatus: HttpStatus,
            headers: HttpHeaders,
            body: String?
        ): RestServiceResponse =
            RestServiceResponse(
                request.operationUrl,
                request.requestEnvelope,
                httpStatus,
                headers,
                body
            )

        fun error(
            request: RestServiceRequest,
            error: Throwable
        ): RestServiceResponse =
            RestServiceResponse(
                request.operationUrl,
                request.requestEnvelope,
                HttpStatus.INTERNAL_SERVER_ERROR,
                HttpHeaders.EMPTY,
                "Ошибка ${error.message} при выполнении запроса к сервису ${request.uri}"
            )

        fun unsupportedMethod(
            request: RestServiceRequest
        ): RestServiceResponse =
            RestServiceResponse(
                request.operationUrl,
                request.requestEnvelope,
                HttpStatus.INTERNAL_SERVER_ERROR,
                HttpHeaders.EMPTY,
                "Ошибка при выполнении запроса к сервису ${request.uri}: метод ${request.method} не поддерживается"
            )
    }
}