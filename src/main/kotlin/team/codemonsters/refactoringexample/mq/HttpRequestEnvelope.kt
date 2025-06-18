package team.codemonsters.refactoringexample.mq

import org.slf4j.LoggerFactory
import team.codemonsters.refactoringexample.mq.enum.OperationType
import team.codemonsters.refactoringexample.mq.enum.getOperationType
import team.codemonsters.refactoringexample.exceptions.ServiceCommunicationException
import team.codemonsters.refactoringexample.util.JsonUtil
import java.time.Duration
import java.time.OffsetDateTime


data class HttpRequestEnvelope private constructor(

    /** Нужно сгенерить на стороне вызывающей стороны*/
    val correlationId: String,

    /** POST - другие методы пока не работают TODO - добавить GET */
    val method: String,

    /** для обработки - cifrub-processing, для данных - cifrub-client-data*/
    val service: String,

    /** Название операции, например customer-buy-cr
     * Список операций для каждого сервиса предопределён и задаётся в конфигурации перекладчика */
    val operation: String,

    /** RequestId (пока наличие не проверяется, но видимо будет) */
    val headers: Map<String, String>,

    /** JSON в виде String */
    val body: String?
) {

    fun toJsonString(): String =
        JsonUtil.toJson(this)

    companion object {

        fun fromJsonString(jsonString: String) =
            Result.runCatching { JsonUtil.fromString(jsonString, HttpRequestEnvelope::class.java) }


        fun emerge(
            correlationId: String,
            httpMethod: String,
            service: String,
            operation: String,
            httpHeaders: Map<String, String>,
            httpBody: String
        ): Result<HttpRequestEnvelope> =
            Result.success(
                HttpRequestEnvelope(
                    correlationId = correlationId,
                    method = httpMethod,
                    service = service,
                    operation = operation,
                    headers = httpHeaders,
                    body = httpBody
                )
            )

        fun putPsRequestInBody(data: Result<HttpRequestEnvelope>): Result<HttpRequestEnvelope> =
            data.fold(
                onSuccess = { request ->
                    emerge(
                        request.correlationId,
                        request.method,
                        request.service,
                        request.operation,
                        request.headers,
                        JsonUtil.toJson(GetWalletBalanceRequest.emerge(request.body!!))
                    )
                },
                onFailure = { error -> Result.failure(error) }
            )


        fun validatedRequest(data: Result<HttpRequestEnvelope>, timeout: Long): Result<HttpRequestEnvelope> =
            data.fold(
                onSuccess = { success -> validatePsRequest(success, timeout) },
                onFailure = { error -> Result.failure(error) }
            )

        private val log = LoggerFactory.getLogger(HttpRequestEnvelope::class.java)!!

        private fun validatePsRequest(data: HttpRequestEnvelope, timeout: Long): Result<HttpRequestEnvelope> =
            PsGetBalanceRequest.fromJsonString(data.body!!).fold(
                onSuccess = { success ->
                    if (!success.operTime.plus(Duration.ofMillis(timeout)).isBefore(OffsetDateTime.now())) {
                        return Result.failure(
                            ServiceCommunicationException("validation", "Истек скорок действия запроса")
                        )
                    }
                    if (OperationType.OPERATION_HISTORY == getOperationType(success.requestType)
                        && success.dateFrom.isBefore(success.dateTo ?: OffsetDateTime.now())
                    ) {
                        return Result.success(data)
                    } else {
                        return Result.failure(
                            ServiceCommunicationException("validation", "Ошибка при валидации входных данных")
                        )
                    }
                },
                onFailure = {
                    log.info("failure")
                    Result.failure(
                        ServiceCommunicationException("validation", "Ошибка при валидации входных данных")
                    )
                }
            )


    }
}


