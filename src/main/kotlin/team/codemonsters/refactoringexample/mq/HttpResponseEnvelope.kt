package team.codemonsters.refactoringexample.mq

import org.springframework.http.HttpHeaders
import org.springframework.http.HttpMethod
import org.springframework.http.HttpStatus
import org.springframework.http.ResponseEntity
import team.codemonsters.refactoringexample.util.JsonUtil

class HttpResponseEnvelope(

    /** Из хедера CorrelationId */
    val correlationId: String,

    /** HttpCode, 500 если ошибка перекладчика */
    val status: Int,

    /** Если была ошибка */
    val message: String?,

    /** Из HttpRequestEnvelope, null если вызова не было*/
    val method: HttpMethod?,

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

    fun toJsonString(): String =
        JsonUtil.toJson(this)


    companion object {
        fun fromJsonString(jsonString: String) =
            Result.runCatching { JsonUtil.fromString(jsonString, HttpResponseEnvelope::class.java) }


        fun emerge(response: RestServiceResponse): Result<HttpResponseEnvelope> {
            return Result.success(
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
            )
        }

        fun emerge(
            data: Result<RestServiceResponse>,
        ): Result<HttpResponseEnvelope> =
            data.fold(
                onSuccess = { restResponse ->
                    Result.success(
                        HttpResponseEnvelope(
                            correlationId = restResponse.requestEnvelope.correlationId,
                            status = restResponse.httpCode.value(),
                            message = restResponse.httpCode.reasonPhrase,
                            method = restResponse.requestEnvelope.method,
                            service = restResponse.requestEnvelope.service,
                            operation = restResponse.requestEnvelope.operation,
                            headers = restResponse.headers,
                            body = restResponse.body ?: ""
                        )
                    )
                },
                onFailure = { error ->
                    Result.failure(error)
                })

    }

    fun emerge(correlationId: String, ex: Throwable): HttpResponseEnvelope {

        return HttpResponseEnvelope(
            correlationId = correlationId,
            status = HttpStatus.INTERNAL_SERVER_ERROR.value(),
            message = HttpStatus.INTERNAL_SERVER_ERROR.reasonPhrase,
            method = null,
            headers = HttpHeaders.EMPTY,
            service = "",
            operation = "",
            body = "Ошибка при процессинге сообщения: ${ex.message}"
        )
    }

}

