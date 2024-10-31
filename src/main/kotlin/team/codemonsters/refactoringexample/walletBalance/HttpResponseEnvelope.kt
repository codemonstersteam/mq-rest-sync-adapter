package team.codemonsters.refactoringexample.walletBalance

import org.springframework.http.HttpHeaders
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import team.codemonsters.refactoringexample.walletBalance.infrastructure.RestServiceResponse

class HttpResponseEnvelope(

    /** Из хедера CorrelationId */
    val correlationId: String,

    /** HttpCode, 500 если ошибка перекладчика */
    val status: Int,

    /** Если была ошибка */
    val message: String?,

    /** Из HttpRequestEnvelope, null если вызова не было*/
    val method: String?,

    /** Из HttpRequestEnvelope */
    val service: String?,

    /** Операция из HttpRequestEnvelope*/
    val operation: String?,

    /** Ответ сервиса в String */
    val body: String?,

    /** Хедеры, которые вернул сервис */
    val headers: HttpHeaders?,
) {
    fun toResponseEntity(): ResponseEntity<String> =
        ResponseEntity<String>(body, HttpStatus.ACCEPTED)

    companion object {

        fun of(response: RestServiceResponse): HttpResponseEnvelope =
            HttpResponseEnvelope(
                correlationId = response.requestEnvelope.correlationId,
                status = response.httpCode.value(),
                message = response.httpCode.reasonPhrase,
                method = response.requestEnvelope.method,
                service = response.requestEnvelope.service,
                operation = response.requestEnvelope.operation,
                headers = response.headers,
                body = response.body ?: ""
            )
    }

}

