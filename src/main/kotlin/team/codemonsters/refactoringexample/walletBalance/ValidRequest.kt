package team.codemonsters.refactoringexample.walletBalance

import team.codemonsters.refactoringexample.exceptions.ServiceCommunicationException
import java.time.Duration
import java.time.OffsetDateTime

data class ValidRequest private constructor(
    val unvalidatedRequest: UnvalidatedRequest,
    val balanceRequest: PsGetBalanceRequest,
) {
    companion object {
        fun emerge(
            unvalidatedRequest: UnvalidatedRequest,
            timeout: Long,
        ): Result<ValidRequest> {
            val httpRequestEnvelope = unvalidatedRequest.httpRequestEnvelope
            if (!unvalidatedRequest.restConfiguration.configs.containsKey(httpRequestEnvelope.service)) {
                return Result.failure(
                    ServiceCommunicationException(
                        unvalidatedRequest.httpRequestEnvelope.service,
                        "Не найдена конфигурация для сервиса: ${httpRequestEnvelope.service}",
                    ),
                )
            }
            if (!unvalidatedRequest.restConfiguration.configs[httpRequestEnvelope.service]!!.operations.containsKey(
                    httpRequestEnvelope.operation,
                )
            ) {
                return Result.failure(
                    ServiceCommunicationException(
                        httpRequestEnvelope.service,
                        "Не найдена операция: ${httpRequestEnvelope.operation} для сервиса ${httpRequestEnvelope.service}",
                    ),
                )
            }
            return PsGetBalanceRequest.fromJsonString(httpRequestEnvelope.body!!)
                .ensure(
                    { OffsetDateTime.now().plus(Duration.ofMillis(timeout)).isBefore(it.operTime) },
                    "Истек срок действия запроса",
                )
                .ensure(
                    { "operationHistory" == it.requestType },
                    "Ошибка при валидации входных данных. Не верный тип запроса.",
                )
                .ensure(
                    { it.dateFrom.isBefore(it.dateTo ?: OffsetDateTime.now()) },
                    "Ошибка при валидации входных данных. Параметр dateFrom должен быть раньше dateTo",
                )
                .map { ValidRequest(unvalidatedRequest, it) }
        }
    }
}
